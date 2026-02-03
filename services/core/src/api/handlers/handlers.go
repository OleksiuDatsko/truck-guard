package handlers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/truckguard/core/src/logic"
	"github.com/truckguard/core/src/models"
	"github.com/truckguard/core/src/repository"
	"github.com/truckguard/core/src/utils"
	"go.opentelemetry.io/otel/trace"
)

func HandlePlateEvent(c *gin.Context) {
	var event models.RawPlateEvent

	if err := c.ShouldBindBodyWith(&event, binding.JSON); err != nil {
		log.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if sysEventID, exists := c.Get("system_event_id"); exists {
		event.SystemEventID = sysEventID.(uint)
	}

	if err := repository.DB.WithContext(c.Request.Context()).Create(&event).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save event"})
		return
	}

	detachCtx := trace.ContextWithSpan(context.Background(), trace.SpanFromContext(c.Request.Context()))
	go logic.MatchPlateEvent(detachCtx, &event)

	c.JSON(http.StatusAccepted, gin.H{"status": "processing", "id": event.ID})
}

func HandlePatchPlateEvent(c *gin.Context) {
	id := c.Param("id")
	var input struct {
		PlateCorrected string `json:"plate_corrected"`
	}
	if err := c.BindJSON(&input); err != nil {
		c.Status(400)
		return
	}
	var event models.RawPlateEvent
	if err := repository.DB.WithContext(c.Request.Context()).First(&event, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
		return
	}

	oldEffectivePlate := event.Plate
	if event.PlateCorrected != "" {
		oldEffectivePlate = event.PlateCorrected
	}

	userID := c.GetHeader("X-User-ID")
	if err := repository.DB.WithContext(c.Request.Context()).Model(&models.RawPlateEvent{}).Where("id = ?", id).Updates(map[string]interface{}{
		"plate_corrected": input.PlateCorrected,
		"corrected_by":    userID,
		"is_manual":       true,
	}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update event"})
		return
	}

	var permits []models.Permit
	err := repository.DB.WithContext(c.Request.Context()).Joins("JOIN permit_plate_events ON permit_plate_events.permit_id = permits.id").
		Where("permit_plate_events.raw_plate_event_id = ?", id).
		Find(&permits).Error

	if err == nil {
		for _, permit := range permits {
			updated := false
			if permit.PlateFront == oldEffectivePlate {
				permit.PlateFront = input.PlateCorrected
				updated = true
			}
			if permit.PlateBack == oldEffectivePlate {
				permit.PlateBack = input.PlateCorrected
				updated = true
			}
			if updated {
				repository.DB.WithContext(c.Request.Context()).Save(&permit)
			}
		}
	}
	c.Status(200)
}

func HandleGetPlateEvents(c *gin.Context) {
	var events []models.RawPlateEvent
	var total int64
	limit, offset, page := utils.GetPagination(c)

	query := repository.DB.WithContext(c.Request.Context()).Model(&models.RawPlateEvent{})

	if plate := c.Query("plate"); plate != "" {
		query = query.Where("plate LIKE ?", "%"+plate+"%")
	}
	if from := c.Query("from"); from != "" {
		query = query.Where("created_at >= ?", from)
	}
	if to := c.Query("to"); to != "" {
		query = query.Where("created_at <= ?", to)
	}

	query.Count(&total)

	if err := query.Limit(limit).Offset(offset).Order("created_at desc").Find(&events).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch events"})
		return
	}
	utils.SendPaginatedResponse(c, events, total, page, limit)
}

func HandleGetPlateEventByID(c *gin.Context) {
	id := c.Param("id")
	var event models.RawPlateEvent
	if err := repository.DB.WithContext(c.Request.Context()).First(&event, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
		return
	}

	var response struct {
		models.RawPlateEvent
		CorrectedByName string `json:"corrected_by_name,omitempty"`
	}
	response.RawPlateEvent = event

	if event.CorrectedBy != "" {
		var user models.User
		if err := repository.DB.WithContext(c.Request.Context()).Where("auth_id = ?", event.CorrectedBy).First(&user).Error; err == nil {
			name := strings.TrimSpace(fmt.Sprintf("%s %s", user.FirstName, user.LastName))
			if name == "" {
				name = fmt.Sprintf("User #%d", user.AuthID)
			}
			response.CorrectedByName = name
		}
	}

	c.JSON(http.StatusOK, response)
}

func HandleWeightEvent(c *gin.Context) {
	var event models.RawWeightEvent

	if err := c.ShouldBindBodyWith(&event, binding.JSON); err != nil {
		log.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if sysEventID, exists := c.Get("system_event_id"); exists {
		event.SystemEventID = sysEventID.(uint)
	}

	if err := repository.DB.WithContext(c.Request.Context()).Create(&event).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to record weight"})
		return
	}

	detachCtx := trace.ContextWithSpan(context.Background(), trace.SpanFromContext(c.Request.Context()))
	go logic.MatchWeightEvent(detachCtx, &event)

	c.JSON(http.StatusAccepted, gin.H{"status": "weight_recorded", "id": event.ID})
}

func HandleGetWeightEvents(c *gin.Context) {
	var events []models.RawWeightEvent
	var total int64
	limit, offset, page := utils.GetPagination(c)

	query := repository.DB.WithContext(c.Request.Context()).Model(&models.RawWeightEvent{})

	if from := c.Query("from"); from != "" {
		query = query.Where("created_at >= ?", from)
	}
	if to := c.Query("to"); to != "" {
		query = query.Where("created_at <= ?", to)
	}

	query.Count(&total)

	if err := query.Limit(limit).Offset(offset).Order("created_at desc").Find(&events).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch weight events"})
		return
	}
	utils.SendPaginatedResponse(c, events, total, page, limit)
}

func HandleGetWeightEventByID(c *gin.Context) {
	id := c.Param("id")
	var event models.RawWeightEvent
	if err := repository.DB.WithContext(c.Request.Context()).First(&event, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Weight event not found"})
		return
	}
	c.JSON(http.StatusOK, event)
}

func HandleGetSystemEvents(c *gin.Context) {
	var events []models.SystemEvent
	var total int64
	limit, offset, page := utils.GetPagination(c)

	query := repository.DB.WithContext(c.Request.Context()).Model(&models.SystemEvent{})

	if eventType := c.Query("type"); eventType != "" {
		query = query.Where("type = ?", eventType)
	}
	if from := c.Query("from"); from != "" {
		query = query.Where("created_at >= ?", from)
	}
	if to := c.Query("to"); to != "" {
		query = query.Where("created_at <= ?", to)
	}

	query.Count(&total)

	if err := query.Limit(limit).Offset(offset).Order("created_at desc").Find(&events).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch system events"})
		return
	}
	utils.SendPaginatedResponse(c, events, total, page, limit)
}

func HandleGetSystemEventByID(c *gin.Context) {
	id := c.Param("id")
	var event models.SystemEvent
	if err := repository.DB.WithContext(c.Request.Context()).First(&event, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "System event not found"})
		return
	}
	c.JSON(http.StatusOK, event)
}

func HandleGetGateEvents(c *gin.Context) {
	var events []models.GateEvent
	var total int64
	limit, offset, page := utils.GetPagination(c)

	query := repository.DB.WithContext(c.Request.Context()).Model(&models.GateEvent{})

	if gate := c.Query("gate"); gate != "" {
		query = query.Where("gate_id = ?", gate)
	}
	if from := c.Query("from"); from != "" {
		query = query.Where("created_at >= ?", from)
	}
	if to := c.Query("to"); to != "" {
		query = query.Where("created_at <= ?", to)
	}

	query.Count(&total)

	if err := query.Limit(limit).Offset(offset).Order("created_at desc").
		Preload("Gate").
		Preload("WeightEvents").Preload("PlateEvents").
		Find(&events).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch gate events"})
		return
	}
	utils.SendPaginatedResponse(c, events, total, page, limit)
}

func HandleGetGateEventByID(c *gin.Context) {
	id := c.Param("id")
	var event models.GateEvent
	if err := repository.DB.WithContext(c.Request.Context()).Preload("Gate").Preload("WeightEvents").Preload("PlateEvents").First(&event, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Gate event not found"})
		return
	}
	c.JSON(http.StatusOK, event)
}
