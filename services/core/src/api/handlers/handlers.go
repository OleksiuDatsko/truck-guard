package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/truckguard/core/src/models"
	"github.com/truckguard/core/src/repository"
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
