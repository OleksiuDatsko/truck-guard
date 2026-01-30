package data

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/truckguard/core/src/models"
	"github.com/truckguard/core/src/repository"
	"github.com/truckguard/core/src/utils"
)

func HandleListCompanies(c *gin.Context) {
	var companies []models.Company
	var total int64
	limit, offset, page := utils.GetPagination(c)
	name := c.Query("name")
	edrpou := c.Query("edrpou")

	query := repository.DB.Model(&models.Company{})
	if name != "" {
		query = query.Where("name ILIKE ?", "%"+name+"%")
	}
	if edrpou != "" {
		query = query.Where("edrpou ILIKE ?", "%"+edrpou+"%")
	}

	query.Count(&total)
	query.Limit(limit).Offset(offset).Order("name asc").Find(&companies)

	utils.SendPaginatedResponse(c, companies, total, page, limit)
}

func HandleCreateCompany(c *gin.Context) {
	var input models.Company
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := repository.DB.Create(&input).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create company"})
		return
	}
	c.JSON(http.StatusCreated, input)
}

func HandleUpdateCompany(c *gin.Context) {
	id := c.Param("id")
	var company models.Company
	if err := repository.DB.First(&company, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Company not found"})
		return
	}
	var input models.Company
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	company.Name = input.Name
	company.EDRPOU = input.EDRPOU
	company.Details = input.Details

	repository.DB.Save(&company)
	c.JSON(http.StatusOK, company)
}

func HandleDeleteCompany(c *gin.Context) {
	id := c.Param("id")
	if err := repository.DB.Delete(&models.Company{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete"})
		return
	}
	c.Status(http.StatusNoContent)
}
