package handlers

import (
	"net/http"
	"strings"

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

	permsHeader := c.GetHeader("X-Permissions")
	hasAllPermits := strings.Contains(permsHeader, "read:permits:all")

	if !hasAllPermits {
		authID := c.GetHeader("X-User-ID")
		var user models.User
		if err := repository.DB.Where("auth_id = ?", authID).First(&user).Error; err != nil {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "User profile not found or access denied"})
			return
		}

		if user.CustomsPostID != nil {
			query = query.Where("customs_post_id = ?", user.CustomsPostID)
		} else {
			c.JSON(http.StatusOK, gin.H{
				"data": []models.Permit{},
				"meta": gin.H{
					"total": 0,
					"page":  page,
					"limit": limit,
				},
			})
			return
		}
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
