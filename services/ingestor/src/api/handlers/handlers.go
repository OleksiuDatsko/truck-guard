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

	camID := c.GetHeader("X-Camera-ID")
	camName := c.GetHeader("X-Camera-Name")

	event, err := repository.ProcessIncomingEvent(file, deviceID, payload, camID, camName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusAccepted, event)
}
