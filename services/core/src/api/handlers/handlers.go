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

	repository.DB.Create(&event)
	c.JSON(http.StatusAccepted, gin.H{"status": "processing"})
}

func HandleWeightEvent(c *gin.Context) {
	var event models.RawWeightEvent

	if err := c.ShouldBindBodyWith(&event, binding.JSON); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	repository.DB.Create(&event)
	c.JSON(http.StatusAccepted, gin.H{"status": "weight_recorded"})
}
