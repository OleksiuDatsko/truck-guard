package data

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/truckguard/core/src/models"
	"github.com/truckguard/core/src/repository"
	"github.com/truckguard/core/src/utils"
)

func HandleListVehicleTypes(c *gin.Context) {
	var types []models.VehicleType
	var total int64
	limit, offset, page := utils.GetPagination(c)
	code := c.Query("code")
	name := c.Query("name")

	query := repository.DB.Model(&models.VehicleType{})
	if code != "" {
		query = query.Where("code ILIKE ?", "%"+code+"%")
	}
	if name != "" {
		query = query.Where("name ILIKE ?", "%"+name+"%")
	}

	query.Count(&total)
	query.Limit(limit).Offset(offset).Order("code asc").Find(&types)

	utils.SendPaginatedResponse(c, types, total, page, limit)
}

func HandleCreateVehicleType(c *gin.Context) {
	var input models.VehicleType
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := repository.DB.Create(&input).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create vehicle type"})
		return
	}
	c.JSON(http.StatusCreated, input)
}

func HandleUpdateVehicleType(c *gin.Context) {
	id := c.Param("id")
	var vehicleType models.VehicleType
	if err := repository.DB.First(&vehicleType, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Vehicle type not found"})
		return
	}
	var input models.VehicleType
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	vehicleType.Name = input.Name
	vehicleType.Code = input.Code
	vehicleType.Description = input.Description
	vehicleType.EntryPrice = input.EntryPrice
	vehicleType.DailyPrice = input.DailyPrice
	vehicleType.Color = input.Color

	repository.DB.Save(&vehicleType)
	c.JSON(http.StatusOK, vehicleType)
}

func HandleDeleteVehicleType(c *gin.Context) {
	id := c.Param("id")
	if err := repository.DB.Delete(&models.VehicleType{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete"})
		return
	}
	c.Status(http.StatusNoContent)
}
