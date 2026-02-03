package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/truckguard/core/src/models"
	"github.com/truckguard/core/src/repository"
	"github.com/truckguard/core/src/utils"
	"gorm.io/datatypes"
)

func HandleGetPermits(c *gin.Context) {
	var permits []models.Permit
	var total int64
	limit, offset, page := utils.GetPagination(c)
	plate := c.Query("plate")

	query := repository.DB.WithContext(c.Request.Context()).Model(&models.Permit{})

	if plate != "" {
		query = query.Where("plate_front = ? OR plate_back = ?", plate, plate)
	}

	permsHeader := c.GetHeader("X-Permissions")
	hasAllPermits := strings.Contains(permsHeader, "read:permits:all")

	if !hasAllPermits {
		authID := c.GetHeader("X-User-ID")
		var user models.User
		if err := repository.DB.WithContext(c.Request.Context()).Where("auth_id = ?", authID).First(&user).Error; err != nil {
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

func HandleCreatePermit(c *gin.Context) {
	var input models.Permit
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	authID := c.GetHeader("X-User-ID")
	// Resolve user ID
	var user models.User
	repository.DB.WithContext(c.Request.Context()).Where("auth_id = ?", authID).First(&user)
	if user.ID != 0 {
		input.CreatedBy = &user.ID
	}

	// 1. Assign Flow
	if input.FlowID == nil {
		// Logic to assign default flow or based on post
		var defaultFlow models.Flow
		if err := repository.DB.WithContext(c.Request.Context()).First(&defaultFlow).Error; err == nil {
			input.FlowID = &defaultFlow.ID
			input.CurrentStepSequence = 1
		}
	} else {
		input.CurrentStepSequence = 1
	}

	input.EntryTime = time.Now()
	input.LastActivityAt = time.Now()

	if err := repository.DB.WithContext(c.Request.Context()).Create(&input).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create permit"})
		return
	}

	// Audit log
	logPermitAudit(c, input.ID, user.ID, "create", map[string]interface{}{"source": "manual"}, "Створено вручну")

	c.JSON(http.StatusCreated, input)
}

func HandleUpdatePermit(c *gin.Context) {
	id := c.Param("id")
	var permit models.Permit
	if err := repository.DB.WithContext(c.Request.Context()).Preload("CustomsPost").First(&permit, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Permit not found"})
		return
	}

	var input map[string]interface{}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	authID := c.GetHeader("X-User-ID")
	var user models.User
	repository.DB.WithContext(c.Request.Context()).Where("auth_id = ?", authID).First(&user)

	// Update fields
	if err := repository.DB.WithContext(c.Request.Context()).Model(&permit).Updates(input).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update permit"})
		return
	}

	// Check verification
	if val, ok := input["verified_by"]; ok && val != nil {
		now := time.Now()
		repository.DB.WithContext(c.Request.Context()).Model(&permit).Update("verified_at", now)
	}

	// Audit

	logPermitAudit(c, permit.ID, user.ID, "update", input, "Оновлено оператором")

	c.JSON(http.StatusOK, permit)
}

func logPermitAudit(c *gin.Context, permitID uint, userID uint, action string, changes interface{}, comment string) {
	jsonBytes, _ := json.Marshal(changes)
	audit := models.PermitAudit{
		PermitID: permitID,
		UserID:   &userID,
		Action:   action,
		Changes:  datatypes.JSON(jsonBytes),
		Comment:  comment,
	}
	repository.DB.WithContext(c.Request.Context()).Create(&audit)
}

func HandleGetPermitByID(c *gin.Context) {
	id := c.Param("id")
	var permit models.Permit
	if err := repository.DB.WithContext(c.Request.Context()).
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
