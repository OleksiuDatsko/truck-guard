package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/truckguard/core/src/models"
	"github.com/truckguard/core/src/repository"
)

func HandleCreateCamera(c *gin.Context) {
	var config models.CameraConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	authURL := "http://auth:8080/admin/keys"
	keyRequest := map[string]interface{}{
		"name":           config.Name + "_key",
		"permission_ids": []string{"create:ingest"},
	}

	jsonData, _ := json.Marshal(keyRequest)

	req, _ := http.NewRequest("POST", authURL, bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-Id", c.GetHeader("X-User-Id"))
	req.Header.Set("X-Permissions", c.GetHeader("X-Permissions"))

	client := &http.Client{}
	resp, err := client.Do(req)

	var authResponse map[string]interface{}
	if err == nil {
		defer resp.Body.Close()
		json.NewDecoder(resp.Body).Decode(&authResponse)
	}

	if idVal, ok := authResponse["id"]; ok {
		config.SourceID = fmt.Sprintf("%v", idVal)
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Auth service did not return an ID"})
		return
	}
	if err := repository.DB.Create(&config).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create camera config"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"camera":  config,
		"api_key": authResponse["api_key"],
	})
}

func HandleGetCameras(c *gin.Context) {
	var configs []models.CameraConfig
	repository.DB.Find(&configs)
	c.JSON(http.StatusOK, configs)
}

func HandleGetConfigByID(c *gin.Context) {
	sourceID := c.Param("id")
	var config models.CameraConfig
	if err := repository.DB.Where("id = ?", sourceID).First(&config).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Camera config not found"})
		return
	}
	c.JSON(http.StatusOK, config)
}

func HandleGetConfigByCameraID(c *gin.Context) {
	sourceID := c.Param("camera_id")
	var config models.CameraConfig
	if err := repository.DB.Where("camera_id = ?", sourceID).First(&config).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Camera config not found"})
		return
	}
	c.JSON(http.StatusOK, config)
}

func HandleUpdateCamera(c *gin.Context) {
	id := c.Param("id")
	var config models.CameraConfig

	if err := repository.DB.First(&config, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Camera configuration not found"})
		return
	}

	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := repository.DB.Save(&config).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update configuration"})
		return
	}

	c.JSON(http.StatusOK, config)
}

func HandleDeleteCamera(c *gin.Context) {
	id := c.Param("id")
	repository.DB.Delete(&models.CameraConfig{}, id)
	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}
