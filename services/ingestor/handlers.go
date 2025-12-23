package main

import (
	"net/http"
	"github.com/gin-gonic/gin"
)

func HandleIngest(c *gin.Context) {
	file, err := c.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "image required"})
		return
	}

	deviceID := c.PostForm("device_id")
	payload := c.PostForm("payload")

	event, err := ProcessIncomingEvent(file, deviceID, payload)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusAccepted, event)
}