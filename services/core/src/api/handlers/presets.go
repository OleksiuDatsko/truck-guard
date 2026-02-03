package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/truckguard/core/src/models"
	"github.com/truckguard/core/src/repository"
	"github.com/truckguard/core/src/utils"
)

func HandleListPresets(c *gin.Context) {
	var presets []models.CameraPreset
	var total int64
	repository.DB.WithContext(c.Request.Context()).Model(&models.CameraPreset{}).Count(&total)

	limit, offset, page := utils.GetPagination(c)
	repository.DB.WithContext(c.Request.Context()).Limit(limit).Offset(offset).Find(&presets)
	utils.SendPaginatedResponse(c, presets, total, page, limit)
}

func HandleGetPreset(c *gin.Context) {
	id := c.Param("id")
	var preset models.CameraPreset
	if err := repository.DB.WithContext(c.Request.Context()).First(&preset, id).Error; err != nil {
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
	if err := repository.DB.WithContext(c.Request.Context()).Create(&preset).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create preset"})
		return
	}
	c.JSON(http.StatusCreated, preset)
}

func HandleUpdatePreset(c *gin.Context) {
	id := c.Param("id")
	var preset models.CameraPreset

	if err := repository.DB.WithContext(c.Request.Context()).First(&preset, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Preset not found"})
		return
	}

	if err := c.ShouldBindJSON(&preset); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	repository.DB.WithContext(c.Request.Context()).Save(&preset)
	c.JSON(http.StatusOK, preset)
}

func HandleDeletePreset(c *gin.Context) {
	id := c.Param("id")
	var count int64
	repository.DB.WithContext(c.Request.Context()).Model(&models.CameraConfig{}).Where("preset_id = ?", id).Count(&count)
	if count > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "Cannot delete preset: it is used by active cameras"})
		return
	}

	if err := repository.DB.WithContext(c.Request.Context()).Delete(&models.CameraPreset{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Delete failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}
