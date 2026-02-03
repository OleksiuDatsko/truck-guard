package data

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/truckguard/core/src/models"
	"github.com/truckguard/core/src/repository"
	"github.com/truckguard/core/src/utils"
)

func HandleListPaymentTypes(c *gin.Context) {
	var types []models.PaymentType
	var total int64
	limit, offset, page := utils.GetPagination(c)
	code := c.Query("code")
	isActive := c.Query("is_active")

	query := repository.DB.WithContext(c.Request.Context()).Model(&models.PaymentType{})
	if code != "" {
		query = query.Where("code ILIKE ?", "%"+code+"%")
	}
	if isActive != "" {
		query = query.Where("is_active = ?", isActive == "true")
	}

	query.Count(&total)
	query.Limit(limit).Offset(offset).Order("code asc").Find(&types)

	utils.SendPaginatedResponse(c, types, total, page, limit)
}

func HandleCreatePaymentType(c *gin.Context) {
	var input models.PaymentType
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := repository.DB.WithContext(c.Request.Context()).Create(&input).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create payment type"})
		return
	}
	c.JSON(http.StatusCreated, input)
}

func HandleUpdatePaymentType(c *gin.Context) {
	id := c.Param("id")
	var paymentType models.PaymentType
	if err := repository.DB.WithContext(c.Request.Context()).First(&paymentType, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Payment type not found"})
		return
	}
	var input models.PaymentType
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	paymentType.Name = input.Name
	paymentType.Code = input.Code
	paymentType.Description = input.Description
	paymentType.IsActive = input.IsActive
	paymentType.Icon = input.Icon

	repository.DB.WithContext(c.Request.Context()).Save(&paymentType)
	c.JSON(http.StatusOK, paymentType)
}

func HandleDeletePaymentType(c *gin.Context) {
	id := c.Param("id")
	if err := repository.DB.WithContext(c.Request.Context()).Delete(&models.PaymentType{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete"})
		return
	}
	c.Status(http.StatusNoContent)
}
