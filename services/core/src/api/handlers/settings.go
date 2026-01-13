package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/truckguard/core/src/models"
	"github.com/truckguard/core/src/repository"
)

// System Settings Handlers

func HandleListSettings(c *gin.Context) {
	var settings []models.SystemSetting
	if err := repository.DB.Find(&settings).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch settings"})
		return
	}
	c.JSON(http.StatusOK, settings)
}

func HandleUpdateSetting(c *gin.Context) {
	var input models.SystemSetting
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var setting models.SystemSetting
	if err := repository.DB.Where("key = ?", input.Key).First(&setting).Error; err != nil {
		// Create if not exists
		setting = input
		if err := repository.DB.Create(&setting).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create setting"})
			return
		}
	} else {
		setting.Value = input.Value
		if err := repository.DB.Save(&setting).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update setting"})
			return
		}
	}

	c.JSON(http.StatusOK, setting)
}

// Excluded Plates Handlers

func HandleListExcludedPlates(c *gin.Context) {
	var plates []models.ExcludedPlate
	if err := repository.DB.Find(&plates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch excluded plates"})
		return
	}
	c.JSON(http.StatusOK, plates)
}

func HandleCreateExcludedPlate(c *gin.Context) {
	var input models.ExcludedPlate
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := repository.DB.Create(&input).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add plate to ignore list"})
		return
	}

	c.JSON(http.StatusCreated, input)
}

func HandleDeleteExcludedPlate(c *gin.Context) {
	id := c.Param("id")
	if err := repository.DB.Delete(&models.ExcludedPlate{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove plate"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}
