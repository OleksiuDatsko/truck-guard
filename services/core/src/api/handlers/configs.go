package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/truckguard/core/src/api/clients"
	"github.com/truckguard/core/src/models"
	"github.com/truckguard/core/src/repository"
	"github.com/truckguard/core/src/utils"
)

func HandleCreateCamera(c *gin.Context) {
	var config models.CameraConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	authClient := clients.NewAuthClient()
	authResp, err := authClient.CreateApiKey(
		c.Request.Context(),
		config.Name+"_key",
		[]string{"create:ingest"},
		c.GetHeader("Authorization"),
		c.GetHeader("X-Api-Key"),
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create authentication key: " + err.Error()})
		return
	}

	config.SourceID = fmt.Sprintf("%v", authResp.ID)

	if err := repository.DB.Create(&config).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save camera configuration"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"camera":  config,
		"api_key": authResp.APIKey,
	})
}

func HandleGetCameras(c *gin.Context) {
	var configs []models.CameraConfig
	var total int64
	limit, offset, page := utils.GetPagination(c)

	repository.DB.Model(&models.CameraConfig{}).Count(&total)
	repository.DB.Limit(limit).Offset(offset).Find(&configs)

	utils.SendPaginatedResponse(c, configs, total, page, limit)
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
	var config models.CameraConfig
	if err := repository.DB.First(&config, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Camera config not found"})
		return
	}

	if config.SourceID != "" {
		authClient := clients.NewAuthClient()
		err := authClient.DeleteApiKey(
			c.Request.Context(),
			config.SourceID,
			c.GetHeader("Authorization"),
			c.GetHeader("X-Api-Key"),
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete associated API key: " + err.Error()})
			return
		}
	}

	repository.DB.Delete(&config)
	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}

func HandleCreateScale(c *gin.Context) {
	var config models.ScaleConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	authClient := clients.NewAuthClient()
	authResp, err := authClient.CreateApiKey(
		c.Request.Context(),
		config.Name+"_key",
		[]string{"create:ingest"},
		c.GetHeader("Authorization"),
		c.GetHeader("X-Api-Key"),
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create authentication key: " + err.Error()})
		return
	}

	config.SourceID = fmt.Sprintf("%v", authResp.ID)

	if err := repository.DB.Create(&config).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save scale configuration"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"scale":   config,
		"api_key": authResp.APIKey,
	})
}

func HandleGetScales(c *gin.Context) {
	var configs []models.ScaleConfig
	var total int64
	limit, offset, page := utils.GetPagination(c)

	repository.DB.Model(&models.ScaleConfig{}).Count(&total)
	repository.DB.Limit(limit).Offset(offset).Find(&configs)

	utils.SendPaginatedResponse(c, configs, total, page, limit)
}

func HandleGetConfigByScaleID(c *gin.Context) {
	scaleID := c.Param("scale_id")
	var config models.ScaleConfig
	if err := repository.DB.Where("scale_id = ?", scaleID).First(&config).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "scale config not found"})
		return
	}
	c.JSON(http.StatusOK, config)
}

func HandleUpdateScale(c *gin.Context) {
	id := c.Param("id")
	var config models.ScaleConfig

	if err := repository.DB.First(&config, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Scale configuration not found"})
		return
	}

	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := repository.DB.Save(&config).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update scale configuration"})
		return
	}

	c.JSON(http.StatusOK, config)
}

func HandleDeleteScale(c *gin.Context) {
	id := c.Param("id")
	repository.DB.Delete(&models.ScaleConfig{}, id)
	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}
