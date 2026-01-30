package data

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/truckguard/core/src/models"
	"github.com/truckguard/core/src/repository"
	"github.com/truckguard/core/src/utils"
)

func HandleListModes(c *gin.Context) {
	var modes []models.CustomsMode
	var total int64
	limit, offset, page := utils.GetPagination(c)
	code := c.Query("code")
	name := c.Query("name")

	query := repository.DB.Model(&models.CustomsMode{})
	if code != "" {
		query = query.Where("code ILIKE ?", "%"+code+"%")
	}
	if name != "" {
		query = query.Where("name ILIKE ?", "%"+name+"%")
	}

	query.Count(&total)
	query.Limit(limit).Offset(offset).Order("code asc").Find(&modes)

	utils.SendPaginatedResponse(c, modes, total, page, limit)
}

func HandleCreateMode(c *gin.Context) {
	var input models.CustomsMode
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := repository.DB.Create(&input).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create mode"})
		return
	}
	c.JSON(http.StatusCreated, input)
}

func HandleUpdateMode(c *gin.Context) {
	id := c.Param("id")
	var mode models.CustomsMode
	if err := repository.DB.First(&mode, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Mode not found"})
		return
	}
	var input models.CustomsMode
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	mode.Name = input.Name
	mode.Code = input.Code
	mode.Description = input.Description

	repository.DB.Save(&mode)
	c.JSON(http.StatusOK, mode)
}

func HandleDeleteMode(c *gin.Context) {
	id := c.Param("id")
	if err := repository.DB.Delete(&models.CustomsMode{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete"})
		return
	}
	c.Status(http.StatusNoContent)
}
