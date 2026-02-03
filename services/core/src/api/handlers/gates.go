package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/truckguard/core/src/models"
	"github.com/truckguard/core/src/repository"
	"github.com/truckguard/core/src/utils"
)

func HandleCreateGate(c *gin.Context) {
	var gate models.Gate
	if err := c.ShouldBindJSON(&gate); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := repository.DB.WithContext(c.Request.Context()).Create(&gate).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create gate"})
		return
	}

	c.JSON(http.StatusCreated, gate)
}

func HandleGetGates(c *gin.Context) {
	var gates []models.Gate
	var total int64
	limit, offset, page := utils.GetPagination(c)

	repository.DB.WithContext(c.Request.Context()).Model(&models.Gate{}).Count(&total)
	repository.DB.WithContext(c.Request.Context()).Limit(limit).Offset(offset).Find(&gates)

	utils.SendPaginatedResponse(c, gates, total, page, limit)
}

func HandleGetGateByID(c *gin.Context) {
	id := c.Param("id")
	var gate models.Gate
	if err := repository.DB.WithContext(c.Request.Context()).First(&gate, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Gate not found"})
		return
	}
	c.JSON(http.StatusOK, gate)
}

func HandleUpdateGate(c *gin.Context) {
	id := c.Param("id")
	var gate models.Gate
	if err := repository.DB.WithContext(c.Request.Context()).First(&gate, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Gate not found"})
		return
	}

	if err := c.ShouldBindJSON(&gate); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := repository.DB.WithContext(c.Request.Context()).Save(&gate).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update gate"})
		return
	}

	c.JSON(http.StatusOK, gate)
}

func HandleDeleteGate(c *gin.Context) {
	id := c.Param("id")
	repository.DB.WithContext(c.Request.Context()).Delete(&models.Gate{}, id)
	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}
