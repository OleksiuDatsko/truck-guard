package logic

import (
	"log"
	"strconv"
	"time"

	"github.com/truckguard/core/src/models"
	"github.com/truckguard/core/src/repository"
)

func MatchPlateEvent(event *models.RawPlateEvent) {
	var count int64
	repository.DB.Model(&models.ExcludedPlate{}).Where("plate = ?", event.Plate).Count(&count)
	if count > 0 {
		log.Printf("[Matchmaker] Plate %s is excluded, skipping permit creation", event.Plate)
		return
	}

	var camera models.CameraConfig
	if err := repository.DB.Where("camera_id = ?", event.SourceID).First(&camera).Error; err != nil {
		log.Printf("[Matchmaker] Failed to find camera config for ID %s", event.SourceID)
		return
	}

	if camera.GateID == nil {
		log.Printf("[Matchmaker] Camera %s is not associated with any Gate", event.SourceID)
		return
	}

	gateID := *camera.GateID
	window := getMatchWindow()

	var permit models.Permit
	var err error

	currentPlate := event.Plate
	if event.PlateCorrected != "" {
		currentPlate = event.PlateCorrected
	}

	err = repository.DB.Where("gate_id = ? AND is_closed = ? AND (plate_front = ? OR plate_back = ?)",
		gateID, false, currentPlate, currentPlate).First(&permit).Error

	if err != nil {
		cutoff := event.Timestamp.Add(-time.Duration(window) * time.Second)
		err = repository.DB.Where("gate_id = ? AND is_closed = ? AND (plate_front = ? OR plate_back = ?) AND updated_at >= ?",
			gateID, true, currentPlate, currentPlate, cutoff).First(&permit).Error
	}

	if err != nil {
		startTime := event.Timestamp.Add(-time.Duration(window) * time.Second)
		endTime := event.Timestamp.Add(time.Duration(window) * time.Second)
		err = repository.DB.Where("gate_id = ? AND is_closed = ? AND entry_time BETWEEN ? AND ?",
			gateID, false, startTime, endTime).First(&permit).Error
	}

	var createdPermit models.Permit
	if err != nil {
		createdPermit = models.Permit{
			GateID:    gateID,
			EntryTime: event.Timestamp,
			IsClosed:  false,
		}
		createdPermit.PlateFront = currentPlate

		repository.DB.Create(&createdPermit)
		permit = createdPermit
	} else {
		updated := false
		if permit.PlateFront == "" {
			permit.PlateFront = currentPlate
			updated = true
		} else if permit.PlateFront != currentPlate && permit.PlateBack == "" {
			permit.PlateBack = currentPlate
			updated = true
		}

		if updated {
			repository.DB.Save(&permit)
		}
	}

	repository.DB.Model(&permit).Association("PlateEvents").Append(event)
}

func MatchWeightEvent(event *models.RawWeightEvent) {
	var scale models.ScaleConfig
	if err := repository.DB.Where("scale_id = ?", event.ScaleID).First(&scale).Error; err != nil {
		log.Printf("[Matchmaker] Failed to find scale config for ID %s", event.ScaleID)
		return
	}

	if scale.GateID == nil {
		log.Printf("[Matchmaker] Scale %s is not associated with any Gate", event.ScaleID)
		return
	}

	gateID := *scale.GateID

	window := getMatchWindow()
	var permit models.Permit
	startTime := event.Timestamp.Add(-1 * time.Hour)
	endTime := event.Timestamp.Add(time.Duration(window) * time.Second)

	err := repository.DB.Where("gate_id = ? AND is_closed = ? AND entry_time BETWEEN ? AND ?",
		gateID, false, startTime, endTime).Order("entry_time asc").First(&permit).Error

	if err == nil {
		permit.TotalWeight = event.Weight
		permit.IsClosed = true
		repository.DB.Save(&permit)

		repository.DB.Model(&permit).Association("WeightEvents").Append(event)
	} else {
		log.Printf("[Matchmaker] No open permit found for weight event at Gate %d", gateID)
	}
}

func getMatchWindow() int {
	var setting models.SystemSetting
	if err := repository.DB.Where("key = ?", "match_window_seconds").First(&setting).Error; err != nil {
		return 120
	}
	val, err := strconv.Atoi(setting.Value)
	if err != nil {
		return 120
	}
	return val
}
