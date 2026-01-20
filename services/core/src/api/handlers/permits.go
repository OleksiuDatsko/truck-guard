package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/truckguard/core/src/models"
	"github.com/truckguard/core/src/repository"
	"github.com/truckguard/core/src/utils"
)

func HandleGetPermits(c *gin.Context) {
	var permits []models.Permit
	var total int64
	limit, offset, page := utils.GetPagination(c)
	plate := c.Query("plate")

	query := repository.DB.Model(&models.Permit{})

	if plate != "" {
		query = query.Where("plate_front = ? OR plate_back = ?", plate, plate)
	}

	query.Count(&total)

	query.Limit(limit).Offset(offset).Order("created_at desc").
		Preload("GateEvents").
		Preload("GateEvents.Gate").
		Preload("GateEvents.PlateEvents").
		Preload("GateEvents.WeightEvents").
		Find(&permits)

	utils.SendPaginatedResponse(c, permits, total, page, limit)
}

func HandleGetPermitByID(c *gin.Context) {
	id := c.Param("id")
	var permit models.Permit
	if err := repository.DB.
		Preload("GateEvents").
		Preload("GateEvents.Gate").
		Preload("GateEvents.PlateEvents").
		Preload("GateEvents.WeightEvents").
		First(&permit, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Permit not found"})
		return
	}

	c.JSON(http.StatusOK, permit)
}
