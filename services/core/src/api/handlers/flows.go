package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/truckguard/core/src/models"
	"github.com/truckguard/core/src/repository"
)

func HandleListFlows(c *gin.Context) {
	var flows []models.Flow
	if err := repository.DB.Preload("Steps").Find(&flows).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch flows"})
		return
	}
	c.JSON(http.StatusOK, flows)
}

func HandleCreateFlow(c *gin.Context) {
	var flow models.Flow
	if err := c.ShouldBindJSON(&flow); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := repository.DB.Create(&flow).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create flow"})
		return
	}
	c.JSON(http.StatusCreated, flow)
}

func HandleGetFlow(c *gin.Context) {
	id := c.Param("id")
	var flow models.Flow
	if err := repository.DB.Preload("Steps.Gate").First(&flow, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Flow not found"})
		return
	}
	c.JSON(http.StatusOK, flow)
}

func HandleUpdateFlow(c *gin.Context) {
	id := c.Param("id")
	var flow models.Flow
	if err := repository.DB.First(&flow, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Flow not found"})
		return
	}

	var input models.Flow
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update basic fields
	flow.Name = input.Name
	flow.Description = input.Description

	// Steps strategy: Replace all steps
	if len(input.Steps) > 0 {
		// remove old steps
		repository.DB.Where("flow_id = ?", flow.ID).Delete(&models.FlowStep{})
		// add new ones
		flow.Steps = input.Steps
	}

	if err := repository.DB.Save(&flow).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update flow"})
		return
	}
	c.JSON(http.StatusOK, flow)
}

func HandleDeleteFlow(c *gin.Context) {
	id := c.Param("id")
	// Clean up steps first (optional if cascade is set, but safter here)
	repository.DB.Where("flow_id = ?", id).Delete(&models.FlowStep{})

	if err := repository.DB.Delete(&models.Flow{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete flow"})
		return
	}
	c.JSON(http.StatusNoContent, nil)
}
