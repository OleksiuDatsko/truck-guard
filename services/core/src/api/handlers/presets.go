package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/truckguard/core/src/models"
	"github.com/truckguard/core/src/repository"
	"net/http"
)

func HandleListPresets(c *gin.Context) {
	var presets []models.CameraPreset
	repository.DB.Find(&presets)
	c.JSON(http.StatusOK, presets)
}

func HandleGetPreset(c *gin.Context) {
	id := c.Param("id")
	var preset models.CameraPreset
	if err := repository.DB.First(&preset, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Preset not found"})
		return
	}
	c.JSON(http.StatusOK, preset)
}

func HandleCreatePreset(c *gin.Context) {
	var preset models.CameraPreset
	if err := c.ShouldBindJSON(&preset); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := repository.DB.Create(&preset).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create preset"})
		return
	}
	c.JSON(http.StatusCreated, preset)
}

func HandleUpdatePreset(c *gin.Context) {
	id := c.Param("id")
	var preset models.CameraPreset

	if err := repository.DB.First(&preset, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Preset not found"})
		return
	}

	if err := c.ShouldBindJSON(&preset); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	repository.DB.Save(&preset)
	c.JSON(http.StatusOK, preset)
}

func HandleDeletePreset(c *gin.Context) {
	id := c.Param("id")
	var count int64
	repository.DB.Model(&models.CameraConfig{}).Where("preset_id = ?", id).Count(&count)
	if count > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "Cannot delete preset: it is used by active cameras"})
		return
	}

	if err := repository.DB.Delete(&models.CameraPreset{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Delete failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}
