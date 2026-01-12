package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/truckguard/core/src/models"
	"github.com/truckguard/core/src/repository"
	"github.com/truckguard/core/src/utils"
)

func HandlePlateEvent(c *gin.Context) {
	var event models.RawPlateEvent

	if err := c.ShouldBindBodyWith(&event, binding.JSON); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if sysEventID, exists := c.Get("system_event_id"); exists {
		event.SystemEventID = sysEventID.(uint)
	}

	if err := repository.DB.Create(&event).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save event"})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{"status": "processing", "id": event.ID})
}

func HandlePatchPlateEvent(c *gin.Context) {
	id := c.Param("id")
	var input struct {
		PlateCorrected string `json:"plate_corrected"`
	}
	if err := c.BindJSON(&input); err != nil {
		return
	}

	repository.DB.Model(&models.RawPlateEvent{}).Where("id = ?", id).Updates(map[string]interface{}{
		"plate_corrected": input.PlateCorrected,
		"is_manual":       true,
	})
	c.Status(200)
}

func HandleGetPlateEvents(c *gin.Context) {
	var events []models.RawPlateEvent
	limit, offset := utils.GetPagination(c)
	if err := repository.DB.Limit(limit).Offset(offset).Order("created_at desc").Find(&events).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch events"})
		return
	}
	c.JSON(http.StatusOK, events)
}

func HandleGetPlateEventByID(c *gin.Context) {
	id := c.Param("id")
	var event models.RawPlateEvent
	if err := repository.DB.First(&event, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
		return
	}
	c.JSON(http.StatusOK, event)
}

func HandleWeightEvent(c *gin.Context) {
	var event models.RawWeightEvent

	if err := c.ShouldBindBodyWith(&event, binding.JSON); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if sysEventID, exists := c.Get("system_event_id"); exists {
		event.SystemEventID = sysEventID.(uint)
	}

	if err := repository.DB.Create(&event).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to record weight"})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{"status": "weight_recorded", "id": event.ID})
}

func HandleGetWeightEvents(c *gin.Context) {
	var events []models.RawWeightEvent
	limit, offset := utils.GetPagination(c)
	if err := repository.DB.Limit(limit).Offset(offset).Order("created_at desc").Find(&events).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch weight events"})
		return
	}
	c.JSON(http.StatusOK, events)
}

func HandleGetWeightEventByID(c *gin.Context) {
	id := c.Param("id")
	var event models.RawWeightEvent
	if err := repository.DB.First(&event, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Weight event not found"})
		return
	}
	c.JSON(http.StatusOK, event)
}

func HandleGetSystemEvents(c *gin.Context) {
	var events []models.SystemEvent
	limit, offset := utils.GetPagination(c)
	if err := repository.DB.Limit(limit).Offset(offset).Order("created_at desc").Find(&events).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch system events"})
		return
	}
	c.JSON(http.StatusOK, events)
}

func HandleGetSystemEventByID(c *gin.Context) {
	id := c.Param("id")
	var event models.SystemEvent
	if err := repository.DB.First(&event, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "System event not found"})
		return
	}
	c.JSON(http.StatusOK, event)
}
