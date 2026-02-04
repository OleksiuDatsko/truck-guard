package handlers

import (
	"net/http"

	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/truckguard/ingestor/src/repository"
)

type CameraMetadata struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func HandleCameraIngest(c *gin.Context) {
	file, err := c.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "image required"})
		return
	}

	deviceID := c.PostForm("device_id")
	payload := c.PostForm("payload")

	sourceID := c.GetHeader("X-Source-ID")
	sourceName := c.GetHeader("X-Source-Name")

	slog.Debug("Incoming camera event", "device_id", deviceID, "source", sourceName)

	event, err := repository.ProcessIncomingEvent(file, deviceID, payload, sourceID, sourceName, "camera", "camera:raw")
	if err != nil {
		slog.Error("Failed to process camera event", "device_id", deviceID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	slog.Info("Camera event ingested", "device_id", deviceID, "event_type", event.Type)
	c.JSON(http.StatusAccepted, event)
}

func HandleWeightIngest(c *gin.Context) {
	deviceID := c.PostForm("device_id")
	payload := c.PostForm("payload")

	sourceID := c.GetHeader("X-Source-ID")
	sourceName := c.GetHeader("X-Source-Name")

	slog.Debug("Incoming weight event", "device_id", deviceID, "source", sourceName)

	event, err := repository.ProcessIncomingEvent(nil, deviceID, payload, sourceID, sourceName, "weight", "weight:raw")
	if err != nil {
		slog.Error("Failed to process weight event", "device_id", deviceID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	slog.Info("Weight event ingested", "device_id", deviceID, "event_type", event.Type)
	c.JSON(http.StatusAccepted, event)
}
