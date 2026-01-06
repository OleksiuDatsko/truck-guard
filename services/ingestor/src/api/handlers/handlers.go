package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/truckguard/ingestor/src/repository"
)

type CameraMetadata struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func HandleIngest(c *gin.Context) {
	file, err := c.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "image required"})
		return
	}

	deviceID := c.PostForm("device_id")
	payload := c.PostForm("payload")

	sourceID := c.GetHeader("X-Source-ID")
	sourceName := c.GetHeader("X-Source-Name")

	event, err := repository.ProcessIncomingEvent(file, deviceID, payload, sourceID, sourceName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusAccepted, event)
}
