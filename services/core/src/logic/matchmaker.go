package logic

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/truckguard/core/src/models"
	"github.com/truckguard/core/src/repository"
)

func getGateDeviceCount(gateID uint) int64 {
	var countCameras int64
	if err := repository.DB.Model(&models.CameraConfig{}).Where("gate_id = ?", gateID).Count(&countCameras).Error; err != nil {
		log.Printf("Failed to get gate camera count: %v", err)
	}

	var countScales int64
	if err := repository.DB.Model(&models.ScaleConfig{}).Where("gate_id = ?", gateID).Count(&countScales).Error; err != nil {
		log.Printf("Failed to get gate scale count: %v", err)
	}

	return countCameras + countScales
}

func getOrCreateGateEvent(gateID uint) uint {
	ctx := context.Background()
	redisKeyID := fmt.Sprintf("active_gate:%d", gateID)
	redisKeyCount := fmt.Sprintf("active_gate_count:%d", gateID)

	gateEventID, err := repository.RDB.Get(ctx, redisKeyID).Uint64()
	if err != nil {
		log.Printf("Failed to get gate event ID: %v", err)
	}
	count, err := repository.RDB.Get(ctx, redisKeyCount).Int64()
	if err != nil {
		log.Printf("Failed to get gate event count: %v", err)
	}

	if gateEventID > 0 && count > 0 {
		return uint(gateEventID)
	}

	newEvent := models.GateEvent{
		GateID:    gateID,
		Timestamp: time.Now(),
	}
	repository.DB.Create(&newEvent)

	maxDevices := getGateDeviceCount(gateID)

	repository.RDB.Set(ctx, redisKeyID, newEvent.ID, 15*time.Second)
	repository.RDB.Set(ctx, redisKeyCount, maxDevices, 15*time.Second)

	return uint(newEvent.ID)
}

func updateGateEventCount(gateID uint) {
	ctx := context.Background()
	redisKeyCount := fmt.Sprintf("active_gate_count:%d", gateID)
	count, err := repository.RDB.Get(ctx, redisKeyCount).Int64()
	if err != nil {
		log.Printf("Failed to get gate event count: %v", err)
		return
	}
	repository.RDB.Set(ctx, redisKeyCount, count-1, 15*time.Second)
}

func MatchPlateEvent(event *models.RawPlateEvent) {
	db := repository.DB

	if err := db.Preload("Camera").First(&event).Error; err != nil {
		log.Printf("Camera %+v has no GateID", &event.Camera)
		return
	}

	var gate models.Gate
	if err := db.Preload("FlowStep").First(&gate, *event.Camera.GateID).Error; err != nil {
		return
	}
	log.Println("Gate:", gate.Name, *event.Camera.GateID)
	gateEventID := getOrCreateGateEvent(gate.ID)
	updateGateEventCount(gate.ID)

	event.GateEventID = &gateEventID
	db.Save(event)

	// Trigger permit processing
	go ProcessGateEventToPermit(gateEventID)
}

func MatchWeightEvent(event *models.RawWeightEvent) {
	db := repository.DB

	if err := db.Preload("Scale").First(&event).Error; err != nil {
		return
	}
	var gate models.Gate
	if err := db.Preload("FlowStep").First(&gate, *event.Scale.GateID).Error; err != nil {
		return
	}
	gateEventID := getOrCreateGateEvent(gate.ID)
	updateGateEventCount(gate.ID)

	event.GateEventID = &gateEventID
	db.Save(event)

	// Trigger permit processing
	go ProcessGateEventToPermit(gateEventID)
}

func ProcessGateEventToPermit(gateEventID uint) {
	var ge models.GateEvent
	// Preload necessary data
	if err := repository.DB.
		Preload("Gate").
		Preload("Gate.FlowStep").
		Preload("PlateEvents").
		Preload("WeightEvents").
		First(&ge, gateEventID).Error; err != nil {
		log.Printf("Failed to load GateEvent %d: %v", gateEventID, err)
		return
	}
	log.Println("PlateEvents:", len(ge.PlateEvents), "WeightEvents:", len(ge.WeightEvents))

	// 1. Identify the best plate (use corrected if available)
	// We iterate through all plate events to find a matching permit or to pick a plate for creation
	var bestPlate string
	var plateCandidates []string

	for _, pe := range ge.PlateEvents {
		log.Println("PlateEvent:", pe.Plate, pe.PlateCorrected)
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

	log.Println("Best plate:", bestPlate, "Plate candidates:", plateCandidates)

	// If no plates yet (e.g. only weight event), we might skip or try to find permit by other means?
	// For now, if no plate, we can't identify vehicle unless we join by time/gate which is risky.
	// But if we are in the same GateEvent, maybe we already linked it to a permit previously?
	// Let's check if GateEvent already has a PermitID
	var permit models.Permit
	var found bool

	if ge.PermitID != nil {
		if err := repository.DB.First(&permit, *ge.PermitID).Error; err == nil {
			found = true
		}
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
			log.Printf("Failed to create permit: %v", err)
		}
	}

	if !found {
		log.Println("Permit not found")
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
	log.Println("Gate:", ge.Gate)
	if ge.Gate.IsExit {
		permit.IsClosed = true
		now := time.Now()
		permit.ExitTime = &now
		dirty = true
		ge.PermitID = &permit.ID
		repository.DB.Save(&ge)
		log.Println("GateEvent saved:", ge)
		log.Println("Permit saved:", permit)
	}

	if dirty {
		permit.LastActivityAt = time.Now()
		repository.DB.Save(&permit)
	}
}
