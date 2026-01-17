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
}
