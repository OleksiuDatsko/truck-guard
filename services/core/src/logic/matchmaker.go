package logic

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/truckguard/core/src/models"
	"github.com/truckguard/core/src/repository"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

var (
	tracer = otel.Tracer("matchmaker")
	meter  = otel.Meter("matchmaker")

	eventsMatchedCounter, _ = meter.Int64Counter("events_matched_total",
		metric.WithDescription("Number of processed plate/weight events"),
	)
	permitsCreatedCounter, _ = meter.Int64Counter("permits_created_total",
		metric.WithDescription("Number of new permits created"),
	)
)

const (
	defaultPermitTimeout = 60 * time.Second
)

// getPermitTimeout fetches the timeout from SystemSettings or returns default
func getPermitTimeout(ctx context.Context) time.Duration {
	var setting models.SystemSetting
	if err := repository.DB.WithContext(ctx).Where("key = ?", "permit_timeout_seconds").First(&setting).Error; err != nil {
		return defaultPermitTimeout
	}
	sec, err := strconv.Atoi(setting.Value)
	if err != nil {
		return defaultPermitTimeout
	}
	slog.Debug("Permit timeout", "timeout", sec)
	return time.Duration(sec) * time.Second
}

func GetOrCreatePermit(ctx context.Context, customsPostID uint) *models.Permit {
	db := repository.DB.WithContext(ctx)
	key := fmt.Sprintf("active_permit:customs_post:%d", customsPostID)

	permitID, _ := repository.RDB.Get(ctx, key).Uint64()
	var permit models.Permit

	if permitID > 0 {
		// Load permit with events to check completeness
		if err := db.Preload("PlateEvents").Preload("WeightEvents").First(&permit, permitID).Error; err == nil {
			// Basic checks
			timeout := getPermitTimeout(ctx)
			if !permit.IsClosed && !permit.IsVoid && permit.VerifiedBy == nil && time.Since(permit.LastActivityAt) < timeout {

				// Check if it's "full" based on trigger sources
				var triggerCameras []models.CameraConfig
				db.Where("customs_post_id = ? AND trigger_permit_creation = ?", customsPostID, true).Find(&triggerCameras)

				var triggerScales []models.ScaleConfig
				db.Where("customs_post_id = ? AND trigger_permit_creation = ?", customsPostID, true).Find(&triggerScales)

				alreadySentSources := make(map[string]bool)
				for _, pe := range permit.PlateEvents {
					alreadySentSources[pe.CameraID] = true
				}
				for _, we := range permit.WeightEvents {
					alreadySentSources[we.ScaleID] = true
				}

				allTriggered := true
				for _, cam := range triggerCameras {
					if !alreadySentSources[cam.SourceID] {
						allTriggered = false
						break
					}
				}
				if allTriggered {
					for _, scale := range triggerScales {
						if !alreadySentSources[scale.SourceID] {
							allTriggered = false
							break
						}
					}
				}

				if !allTriggered {
					slog.Debug("Reusing active permit", "permit_id", permit.ID)
					return &permit
				}
				slog.Debug("Active permit is full, will create new one", "permit_id", permit.ID)
			}
		}
	}

	// Create new permit
	permit = models.Permit{
		CustomsPostID:  &customsPostID,
		Code:           fmt.Sprintf("PRM-%d-%d", customsPostID, time.Now().Unix()),
		EntryTime:      time.Now(),
		LastActivityAt: time.Now(),
	}

	if err := db.Create(&permit).Error; err != nil {
		slog.Error("Failed to create permit", "error", err)
		return nil
	}

	// Set in Redis with timeout
	timeout := getPermitTimeout(ctx)
	repository.RDB.Set(ctx, key, permit.ID, timeout)

	permitsCreatedCounter.Add(ctx, 1)
	slog.Info("Created new permit", "permit_id", permit.ID, "code", permit.Code)

	return &permit
}

func MatchPlateEvent(ctx context.Context, event *models.PlateEvent) {
	ctx, span := tracer.Start(ctx, "MatchPlateEvent",
		trace.WithAttributes(attribute.String("plate", event.Plate)))
	defer span.End()

	db := repository.DB
	// Reload event to get associations if needed, or assume it's passed with ID.
	// We need Camera to get CustomsPostID.
	if err := db.WithContext(ctx).Preload("Camera").First(event).Error; err != nil {
		slog.Error("Failed to load plate event", "error", err)
		span.RecordError(err)
		return
	}

	if event.Camera.CustomsPostID == nil {
		slog.Warn("Camera not assigned to a CustomsPost", "camera_id", event.CameraID)
		return
	}
	customsPostID := *event.Camera.CustomsPostID

	processEvent(ctx, customsPostID, event, func(permit *models.Permit) {
		plate := event.Plate
		if event.PlateCorrected != "" {
			plate = event.PlateCorrected
		}

		switch event.Camera.Type {
case "front":
			permit.PlateFront = plate
		case "back":
			permit.PlateBack = plate
		default:
			slog.Warn("Unknown camera type", "camera_type", event.Camera.Type)
			if permit.PlateFront == "" {
				permit.PlateFront = plate
			} else if permit.PlateBack == "" && permit.PlateFront != plate {
				permit.PlateBack = plate
			}
		}
	})

	eventsMatchedCounter.Add(ctx, 1, metric.WithAttributes(
		attribute.String("type", "plate"),
	))
}

func MatchWeightEvent(ctx context.Context, event *models.WeightEvent) {
	ctx, span := tracer.Start(ctx, "MatchWeightEvent",
		trace.WithAttributes(attribute.Float64("weight", event.Weight)))
	defer span.End()

	db := repository.DB
	if err := db.WithContext(ctx).Preload("Scale").First(event).Error; err != nil {
		slog.Error("Failed to load weight event", "error", err)
		span.RecordError(err)
		return
	}

	if event.Scale.CustomsPostID == nil {
		slog.Warn("Scale not assigned to a CustomsPost", "scale_id", event.ScaleID)
		return
	}
	customsPostID := *event.Scale.CustomsPostID

	processEvent(ctx, customsPostID, event, func(permit *models.Permit) {
		if event.Weight > 0 {
			permit.TotalWeight = event.Weight
		}
	})

	eventsMatchedCounter.Add(ctx, 1, metric.WithAttributes(
		attribute.String("type", "weight"),
	))
}

// processEvent handles the common logic: find active permit or create new, then apply updates
func processEvent(ctx context.Context, customsPostID uint, event interface{}, updateFn func(*models.Permit)) {
	db := repository.DB.WithContext(ctx)
	permit := GetOrCreatePermit(ctx, customsPostID)
	if permit == nil {
		return
	}

	// Update existing
	updateFn(permit)
	permit.LastActivityAt = time.Now()
	if err := db.Save(permit).Error; err != nil {
		slog.Error("Failed to update permit", "permit_id", permit.ID, "error", err)
		return
	}
	slog.Info("Processed event for permit", "permit_id", permit.ID)

	// Refresh Redis TTL
	key := fmt.Sprintf("active_permit:customs_post:%d", customsPostID)
	timeout := getPermitTimeout(ctx)
	repository.RDB.Set(ctx, key, permit.ID, timeout)

	// Link Event to Permit
	switch e := event.(type) {
	case *models.PlateEvent:
		e.PermitID = &permit.ID
		db.Save(e)
	case *models.WeightEvent:
		e.PermitID = &permit.ID
		db.Save(e)
	}
}
