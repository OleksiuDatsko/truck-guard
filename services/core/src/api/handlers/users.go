package handlers

import (
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/truckguard/core/src/api/clients"
	"github.com/truckguard/core/src/models"
	"github.com/truckguard/core/src/repository"
)

func HandleCreateUser(c *gin.Context) {
	var input struct {
		Username      string `json:"username" binding:"required"`
		Password      string `json:"password" binding:"required"`
		Role          string `json:"role"`
		FirstName     string `json:"first_name"`
		LastName      string `json:"last_name"`
		ThirdName     string `json:"third_name"`
		PhoneNumber   string `json:"phone_number"`
		Email         string `json:"email"`
		Notes         string `json:"notes"`
		CustomsPostID *uint  `json:"customs_post_id"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	authClient := clients.NewAuthClient()
	if authClient == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Auth client not configured"})
		return
	}

	authResp, err := authClient.RegisterUser(
		c.Request.Context(),
		input.Username,
		input.Password,
		input.Role,
		c.GetHeader("Authorization"),
		c.GetHeader("X-Api-Key"),
	)

	if err != nil {
		if strings.Contains(err.Error(), "status: 409") {
			c.JSON(http.StatusConflict, gin.H{"error": "User already exists"})
			return
		}
		c.JSON(http.StatusBadGateway, gin.H{"error": "Auth service registration failed: " + err.Error()})
		return
	}

	user := models.User{
		AuthID:        authResp.ID,
		FirstName:     input.FirstName,
		LastName:      input.LastName,
		ThirdName:     input.ThirdName,
		PhoneNumber:   input.PhoneNumber,
		Email:         input.Email,
		Notes:         input.Notes,
		Role:          input.Role,
		CustomsPostID: input.CustomsPostID,
	}

	if err := repository.DB.WithContext(c.Request.Context()).Create(&user).Error; err != nil {
		slog.Error("Failed to create profile in Core", "auth_id", user.AuthID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create profile in Core"})
		return
	}
	slog.Info("User created successfully", "auth_id", user.AuthID, "username", input.Username)

	c.JSON(http.StatusCreated, user)
}

func HandleListUsers(c *gin.Context) {
	var users []models.User
	if err := repository.DB.WithContext(c.Request.Context()).Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users"})
		return
	}
	c.JSON(http.StatusOK, users)
}

func HandleGetUser(c *gin.Context) {
	id := c.Param("id")
	var user models.User
	if err := repository.DB.WithContext(c.Request.Context()).Where("ID = ?", id).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User profile not found"})
		return
	}
	c.JSON(http.StatusOK, user)
}

func HandleGetUserByAuthID(c *gin.Context) {
	authID := c.Param("authId")
	var user models.User
	if err := repository.DB.WithContext(c.Request.Context()).Where("auth_id = ?", authID).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User profile not found for this Auth ID"})
		return
	}
	c.JSON(http.StatusOK, user)
}

func HandleDeleteUser(c *gin.Context) {
	id := c.Param("id")
	slog.Debug("Processing user deletion request", "auth_id", id)
	var user models.User
	if err := repository.DB.WithContext(c.Request.Context()).Where("auth_id = ?", id).First(&user).Error; err != nil {
		slog.Warn("User profile not found for deletion", "auth_id", id)
		c.JSON(http.StatusNotFound, gin.H{"error": "User profile not found"})
		return
	}
	authClient := clients.NewAuthClient()
	if authClient != nil {
		err := authClient.DeleteUser(
			c.Request.Context(),
			user.AuthID,
			c.GetHeader("Authorization"),
			c.GetHeader("X-Api-Key"),
		)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": "Failed to delete user from Auth service"})
			return
		}
	}

	if err := repository.DB.WithContext(c.Request.Context()).Where("auth_id = ?", id).Delete(&models.User{}).Error; err != nil {
		slog.Error("Failed to delete profile from Core", "auth_id", id, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete profile from Core"})
		return
	}

	slog.Info("User deleted successfully from all services", "auth_id", id)
	c.JSON(http.StatusOK, gin.H{"status": "deleted from both services"})
}

func HandleUpdateUser(c *gin.Context) {
	id := c.Param("id")
	var input struct {
		FirstName     string `json:"first_name"`
		LastName      string `json:"last_name"`
		ThirdName     string `json:"third_name"`
		PhoneNumber   string `json:"phone_number"`
		Email         string `json:"email"`
		Notes         string `json:"notes"`
		CustomsPostID *uint  `json:"customs_post_id"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := repository.DB.WithContext(c.Request.Context()).Where("auth_id = ?", id).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User profile not found"})
		return
	}

	user.FirstName = input.FirstName
	user.LastName = input.LastName
	user.ThirdName = input.ThirdName
	user.PhoneNumber = input.PhoneNumber
	user.Email = input.Email
	user.Notes = input.Notes
	user.CustomsPostID = input.CustomsPostID

	if err := repository.DB.WithContext(c.Request.Context()).Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile in Core"})
		return
	}

	c.JSON(http.StatusOK, user)
}

func HandleGetMyProfile(c *gin.Context) {
	authID := c.GetHeader("X-User-ID")
	if authID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	authUser, err := clients.NewAuthClient().GetUser(c.Request.Context(), authID, c.GetHeader("Authorization"))
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "Failed to get user from Auth service"})
		return
	}

	var user models.User
	if err := repository.DB.WithContext(c.Request.Context()).Where("auth_id = ?", authID).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Profile not found"})
		return
	}

	response := gin.H{
		"id":           user.ID,
		"auth_id":      user.AuthID,
		"first_name":   user.FirstName,
		"last_name":    user.LastName,
		"third_name":   user.ThirdName,
		"phone_number": user.PhoneNumber,
		"email":        user.Email,
		"notes":        user.Notes,
		"username":     authUser.Username,
		"role":         authUser.Role,
	}

	c.JSON(http.StatusOK, response)
}

func HandleUpdateMyProfile(c *gin.Context) {
	authIDStr := c.GetHeader("X-User-ID")
	if authIDStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var authID uint64
	var err error
	if _, err = fmt.Sscanf(authIDStr, "%d", &authID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid User ID format"})
		return
	}
	realAuthID := uint(authID)

	var input struct {
		FirstName   string `json:"first_name"`
		LastName    string `json:"last_name"`
		ThirdName   string `json:"third_name"`
		PhoneNumber string `json:"phone_number"`
		Email       string `json:"email"`
		Notes       string `json:"notes"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	result := repository.DB.WithContext(c.Request.Context()).Where("auth_id = ?", realAuthID).First(&user)

	if result.Error != nil {
		user = models.User{
			AuthID: realAuthID,
		}
	}

	user.FirstName = input.FirstName
	user.LastName = input.LastName
	user.ThirdName = input.ThirdName
	user.PhoneNumber = input.PhoneNumber
	user.Email = input.Email
	user.Notes = input.Notes

	if result.Error != nil {
		if err := repository.DB.WithContext(c.Request.Context()).Create(&user).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create profile"})
			return
		}
	} else {
		if err := repository.DB.WithContext(c.Request.Context()).Save(&user).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile"})
			return
		}
	}

	c.JSON(http.StatusOK, user)
}
