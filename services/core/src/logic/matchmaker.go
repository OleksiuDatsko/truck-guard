package logic

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/truckguard/core/src/models"
	"github.com/truckguard/core/src/repository"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

var tracer = otel.Tracer("matchmaker")

func getGateDeviceCount(gateID uint) int64 {
	var countCameras int64
	if err := repository.DB.Model(&models.CameraConfig{}).Where("gate_id = ?", gateID).Count(&countCameras).Error; err != nil {
		slog.Error("Failed to get gate camera count", "gate_id", gateID, "error", err)
	}

	var countScales int64
	if err := repository.DB.Model(&models.ScaleConfig{}).Where("gate_id = ?", gateID).Count(&countScales).Error; err != nil {
		slog.Error("Failed to get gate scale count", "gate_id", gateID, "error", err)
	}

	return countCameras + countScales
}

func MatchPlateEvent(ctx context.Context, event *models.RawPlateEvent) {
	ctx, span := tracer.Start(ctx, "MatchPlateEvent",
		trace.WithAttributes(attribute.String("plate", event.Plate)))
	defer span.End()

	db := repository.DB

	if err := db.WithContext(ctx).Preload("Camera").First(&event).Error; err != nil {
		slog.Error("Camera has no GateID", "camera_id", event.Camera.ID, "camera", event.Camera)
		span.RecordError(err)
		return
	}

	var gate models.Gate
	if err := db.WithContext(ctx).Preload("FlowStep").First(&gate, *event.Camera.GateID).Error; err != nil {
		span.RecordError(err)
		return
	}
	slog.Info("Gate identified", "gate_name", gate.Name, "gate_id", *event.Camera.GateID)
	gateEventID := getOrCreateGateEvent(ctx, gate.ID)
	updateGateEventCount(ctx, gate.ID)

	event.GateEventID = &gateEventID
	db.WithContext(ctx).Save(event)

	// Trigger permit processing
	go ProcessGateEventToPermit(ctx, gateEventID)
}

func MatchWeightEvent(ctx context.Context, event *models.RawWeightEvent) {
	ctx, span := tracer.Start(ctx, "MatchWeightEvent",
		trace.WithAttributes(attribute.Float64("weight", event.Weight)))
	defer span.End()

	db := repository.DB

	if err := db.WithContext(ctx).Preload("Scale").First(&event).Error; err != nil {
		span.RecordError(err)
		return
	}
	var gate models.Gate
	if err := db.WithContext(ctx).Preload("FlowStep").First(&gate, *event.Scale.GateID).Error; err != nil {
		span.RecordError(err)
		return
	}
	gateEventID := getOrCreateGateEvent(ctx, gate.ID)
	updateGateEventCount(ctx, gate.ID)

	event.GateEventID = &gateEventID
	db.WithContext(ctx).Save(event)

	// Trigger permit processing
	go ProcessGateEventToPermit(ctx, gateEventID)
}

func ProcessGateEventToPermit(ctx context.Context, gateEventID uint) {
	ctx, span := tracer.Start(ctx, "ProcessGateEventToPermit",
		trace.WithAttributes(attribute.Int("gate_event_id", int(gateEventID))))
	defer span.End()

	var ge models.GateEvent
	// Preload necessary data
	if err := repository.DB.
		Preload("Gate").
		Preload("Gate.FlowStep").
		Preload("PlateEvents").
		Preload("WeightEvents").
		First(&ge, gateEventID).Error; err != nil {
		slog.Error("Failed to load GateEvent", "gate_event_id", gateEventID, "error", err)
		return
	}
	slog.Info("Gathered gate events", "plate_events_count", len(ge.PlateEvents), "weight_events_count", len(ge.WeightEvents))

	// 1. Identify the best plate (use corrected if available)
	// We iterate through all plate events to find a matching permit or to pick a plate for creation
	var bestPlate string
	var plateCandidates []string

	for _, pe := range ge.PlateEvents {
		slog.Debug("Processing plate event", "plate", pe.Plate, "plate_corrected", pe.PlateCorrected)
		p := pe.Plate
		if pe.PlateCorrected != "" {
			p = pe.PlateCorrected
		}
		if p != "" {
			plateCandidates = append(plateCandidates, p)
			if bestPlate == "" {
				bestPlate = p
			}
		}
	}

	slog.Info("Plate candidates identified", "best_plate", bestPlate, "candidates", plateCandidates)

	// If no plates yet (e.g. only weight event), we might skip or try to find permit by other means?
	// For now, if no plate, we can't identify vehicle unless we join by time/gate which is risky.
	// But if we are in the same GateEvent, maybe we already linked it to a permit previously?
	// Let's check if GateEvent already has a PermitID
	var permit models.Permit
	var found bool

	// If ge.Gate.FlowStep != nil { ... }
	if ge.PermitID != nil {
		span.AddEvent("linking_gate_event_to_permit", trace.WithAttributes(attribute.Int("permit_id", int(*ge.PermitID))))
		ge.PermitID = &permit.ID
		repository.DB.WithContext(ctx).Save(&ge)
	}

	if !found && len(plateCandidates) > 0 {
		// Search for OPEN permit matching ANY of the candidates
		// We use a query that checks PlateFront OR PlateBack matches any candidate
		err := repository.DB.Where("(plate_front IN ? OR plate_back IN ?) AND is_closed = ?", plateCandidates, plateCandidates, false).
			First(&permit).Error
		if err == nil {
			found = true
			// Link this gate event to the found permit
			ge.PermitID = &permit.ID
			repository.DB.Save(&ge)
		}
	}

	// 2. Entry Gate Logic (Create Permit)
	if !found && ge.Gate.IsEntry && bestPlate != "" {
		// Create new permit
		newPermit := models.Permit{
			PlateFront:          bestPlate,    // Assumption: First seen is Front. TODO: Logic to distinguish Front/Back
			EntryTime:           ge.Timestamp, // Or time.Now()
			IsClosed:            false,
			CurrentStepSequence: 1, // Default start
		}

		// Assign FlowID if Gate is part of a Flow
		if ge.Gate.FlowStep != nil {
			newPermit.FlowID = &ge.Gate.FlowStep.FlowID
			newPermit.CurrentStepSequence = ge.Gate.FlowStep.Sequence
		}

		if err := repository.DB.Create(&newPermit).Error; err == nil {
			permit = newPermit
			found = true
			// Link GateEvent
			ge.PermitID = &permit.ID
			repository.DB.Save(&ge)
		} else {
			slog.Error("Failed to create permit", "best_plate", bestPlate, "error", err)
		}
	}

	if !found {
		slog.Warn("Permit not found for gate event", "gate_event_id", ge.ID, "plate_candidates", plateCandidates)
		return
	}

	// 3. Update Permit Data
	dirty := false

	// Update Weight
	if len(ge.WeightEvents) > 0 {
		// Take the last weight or max weight? User said "WeightEvents[0]" but we might have multiple.
		// Let's take the latest or just the first one.
		w := ge.WeightEvents[0].Weight
		if w > 0 {
			permit.TotalWeight = w
			dirty = true
		}
	}

	// Update Plates if missing (Simple logic)
	if permit.PlateBack == "" && len(plateCandidates) > 1 {
		// If we have more than one distinct plate, maybe the other is back?
		// SImplification: If bestPlate is assigned to PlateFront, assign next candidate to PlateBack?
		// Or assume PlateBack comes from "Back" camera?
		// RawPlateEvent has CameraSourceName?
		// For now, if we have a second distinct plate, assign it.
		for _, p := range plateCandidates {
			if p != permit.PlateFront {
				permit.PlateBack = p
				dirty = true
				break
			}
		}
	}

	// Validate/Update Sequence (Mid/Exit)
	// If the Gate has a sequence number, we update CurrentStepSequence
	if ge.Gate.FlowStep != nil {
		// Check invalid jumps?
		// For now, simple update
		if ge.Gate.FlowStep.Sequence > permit.CurrentStepSequence {
			permit.CurrentStepSequence = ge.Gate.FlowStep.Sequence
			dirty = true
		}
	}

	// 4. Exit Gate Logic
	slog.Debug("Checking exit gate", "gate_id", ge.Gate.ID, "is_exit", ge.Gate.IsExit)
	if ge.Gate.IsExit {
		permit.IsClosed = true
		now := time.Now()
		permit.ExitTime = &now
		dirty = true
		ge.PermitID = &permit.ID
		repository.DB.WithContext(ctx).Save(&ge)
		slog.Debug("GateEvent saved", "gate_event_id", ge.ID)
		slog.Debug("Permit saved", "permit_id", permit.ID, "is_closed", permit.IsClosed)
	}

	if dirty {
		permit.LastActivityAt = time.Now()
		repository.DB.WithContext(ctx).Save(&permit)
	}
}

func getOrCreateGateEvent(ctx context.Context, gateID uint) uint {
	ctx, span := tracer.Start(ctx, "getOrCreateGateEvent")
	defer span.End()

	redisKeyID := fmt.Sprintf("active_gate:%d", gateID)
	redisKeyCount := fmt.Sprintf("active_gate_count:%d", gateID)

	gateEventID, err := repository.RDB.Get(ctx, redisKeyID).Uint64()
	if err != nil {
		slog.Debug("Failed to get active gate event ID from Redis", "gate_id", gateID, "error", err)
	}
	count, err := repository.RDB.Get(ctx, redisKeyCount).Int64()
	if err != nil {
		slog.Debug("Failed to get active gate event count from Redis", "gate_id", gateID, "error", err)
	}

	if gateEventID > 0 && count > 0 {
		return uint(gateEventID)
	}

	newEvent := models.GateEvent{
		GateID:    gateID,
		Timestamp: time.Now(),
	}
	repository.DB.WithContext(ctx).Create(&newEvent)

	maxDevices := getGateDeviceCount(gateID)

	repository.RDB.Set(ctx, redisKeyID, newEvent.ID, 15*time.Second)
	repository.RDB.Set(ctx, redisKeyCount, maxDevices, 15*time.Second)

	return uint(newEvent.ID)
}

func updateGateEventCount(ctx context.Context, gateID uint) {
	redisKeyCount := fmt.Sprintf("active_gate_count:%d", gateID)
	count, err := repository.RDB.Get(ctx, redisKeyCount).Int64()
	if err != nil {
		slog.Error("Failed to get gate event count from Redis for update", "gate_id", gateID, "error", err)
		return
	}
	repository.RDB.Set(ctx, redisKeyCount, count-1, 15*time.Second)
}
