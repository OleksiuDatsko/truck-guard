package handlers

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/truckguard/auth/src/models"
	"github.com/truckguard/auth/src/repository"
	"golang.org/x/crypto/bcrypt"
)

func HandleRegister(c *gin.Context) {
	var b struct {
		User string `json:"username"`
		Pass string `json:"password"`
		Role string `json:"role"`
	}
	if err := c.BindJSON(&b); err != nil {
		c.Status(400)
		return
	}

	h, _ := bcrypt.GenerateFromPassword([]byte(b.Pass), 10)

	var role models.Role
	roleName := b.Role
	if roleName == "" {
		roleName = "operator"
	}
	repository.DB.Where("name = ?", roleName).First(&role)

	u := models.User{Username: b.User, PasswordHash: string(h), RoleID: role.ID}
	if err := repository.DB.Preload("Role").Preload("Role.Permissions").Create(&u).Error; err != nil {
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			c.JSON(409, gin.H{"error": "User already exists"})
			return
		}
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(201, u)
}

func HandleLogin(c *gin.Context) {
	var b struct {
		User string `json:"username"`
		Pass string `json:"password"`
	}
	if err := c.BindJSON(&b); err != nil {
		c.Status(400)
		return
	}

	var u models.User
	if err := repository.DB.Preload("Role.Permissions").Where("username = ?", b.User).First(&u).Error; err != nil {
		c.Status(401)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(b.Pass)); err != nil {
		c.Status(401)
		return
	}

	now := time.Now()
	u.LastLogin = &now
	repository.DB.Save(&u)

	t, _ := repository.GenerateToken(u)
	c.JSON(200, gin.H{"token": t})
}

func HandleValidate(c *gin.Context) {
	k := c.GetHeader("X-API-Key")
	if k != "" {
		if meta, valid := repository.ValidateKeyAndGetMetadata(k); valid {
			c.Header("X-Source-ID", meta.ID)
			c.Header("X-Source-Name", meta.Name)
			c.Header("X-Permissions", strings.Join(meta.Permissions, ","))
			c.Status(200)
			return
		}
	}

	a := c.GetHeader("Authorization")
	log.Println(a)
	if strings.HasPrefix(a, "Bearer ") {
		ts := strings.TrimPrefix(a, "Bearer ")
		token, err := jwt.Parse(ts, func(t *jwt.Token) (interface{}, error) { return repository.JWTSecret, nil })

		if err == nil && token.Valid {
			if claims, ok := token.Claims.(jwt.MapClaims); ok {
				userID := uint(claims["user_id"].(float64))

				perms := repository.GetUserPermissions(userID)

				c.Header("X-Permissions", strings.Join(perms, ","))
				c.Header("X-User-ID", fmt.Sprintf("%d", userID))

				c.Status(200)
				return
			}
		}
	}
	c.Status(401)
}

func HandleListPermissions(c *gin.Context) {
	var perms []models.Permission
	repository.DB.Find(&perms)
	c.JSON(200, perms)
}

func HandleCreateRole(c *gin.Context) {
	var b struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if err := c.BindJSON(&b); err != nil {
		c.Status(400)
		return
	}

	role := models.Role{Name: b.Name, Description: b.Description}
	if err := repository.DB.Create(&role).Error; err != nil {
		c.JSON(500, gin.H{"error": "Role already exists or DB error"})
		return
	}
	c.JSON(201, role)
}

func HandleAssignPermissionsToRole(c *gin.Context) {
	roleID := c.Param("id")
	var b struct {
		PermissionIDs []string `json:"permission_ids"`
	}
	if err := c.BindJSON(&b); err != nil {
		c.Status(400)
		return
	}

	var role models.Role
	if err := repository.DB.First(&role, roleID).Error; err != nil {
		c.JSON(404, gin.H{"error": "Role not found"})
		return
	}

	var perms []models.Permission
	repository.DB.Where("id IN ?", b.PermissionIDs).Find(&perms)

	if err := repository.DB.Model(&role).Association("Permissions").Replace(perms); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	var userIDs []uint
	repository.DB.Model(&models.User{}).Where("role_id = ?", role.ID).Pluck("id", &userIDs)
	for _, id := range userIDs {
		repository.InvalidateUserCache(id)
	}

	c.JSON(200, gin.H{"message": "Permissions updated for role " + role.Name})
}

func HandleUpdateUserRole(c *gin.Context) {
	userIDStr := c.Param("id")
	var b struct {
		RoleID uint `json:"role_id"`
	}
	if err := c.BindJSON(&b); err != nil {
		c.Status(400)
		return
	}

	repository.DB.Model(&models.User{}).Where("id = ?", userIDStr).Update("role_id", b.RoleID)

	var id uint
	fmt.Sscanf(userIDStr, "%d", &id)
	repository.InvalidateUserCache(id)

	c.Status(200)
}

func HandleListKeys(c *gin.Context) {
	var keys []models.APIKey
	repository.DB.Preload("Permissions").Find(&keys)
	c.JSON(200, keys)
}

func HandleCreateKeyWithPerms(c *gin.Context) {
	var b struct {
		Name          string   `json:"name"`
		PermissionIDs []string `json:"permission_ids"`
	}
	if err := c.BindJSON(&b); err != nil {
		c.Status(400)
		return
	}

	rb := make([]byte, 16)
	rand.Read(rb)
	rk := hex.EncodeToString(rb)

	var perms []models.Permission
	repository.DB.Where("id IN ?", b.PermissionIDs).Find(&perms)

	key := models.APIKey{
		KeyHash:     repository.HashKey(rk),
		OwnerName:   b.Name,
		Permissions: perms,
	}

	if err := repository.DB.Create(&key).Error; err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(201, gin.H{"api_key": rk, "id": key.ID})
}

func HandleUpdateKeyStatus(c *gin.Context) {
	id := c.Param("id")
	var b struct {
		IsActive bool `json:"is_active"`
	}
	if err := c.BindJSON(&b); err != nil {
		c.Status(400)
		return
	}

	var key models.APIKey
	if err := repository.DB.First(&key, id).Error; err != nil {
		c.JSON(404, gin.H{"error": "Key not found"})
		return
	}

	repository.DB.Model(&key).Update("is_active", b.IsActive)

	repository.RDB.Del(context.Background(), "auth:"+key.KeyHash)

	c.Status(200)
}

func HandleAssignPermissionsToKey(c *gin.Context) {
	id := c.Param("id")
	var b struct {
		PermissionIDs []string `json:"permission_ids"`
	}
	if err := c.BindJSON(&b); err != nil {
		c.Status(400)
		return
	}

	var key models.APIKey
	if err := repository.DB.First(&key, id).Error; err != nil {
		c.JSON(404, gin.H{"error": "Key not found"})
		return
	}

	var perms []models.Permission
	repository.DB.Where("id IN ?", b.PermissionIDs).Find(&perms)

	repository.DB.Model(&key).Association("Permissions").Replace(perms)

	repository.RDB.Del(context.Background(), "auth:"+key.KeyHash)

	c.JSON(200, gin.H{"message": "Permissions updated for key " + key.OwnerName})
}

func HandleUpdateKey(c *gin.Context) {
	id := c.Param("id")
	var b struct {
		OwnerName string `json:"owner_name"`
		IsActive  bool   `json:"is_active"`
	}
	if err := c.BindJSON(&b); err != nil {
		c.Status(400)
		return
	}

	var key models.APIKey
	if err := repository.DB.First(&key, id).Error; err != nil {
		c.JSON(404, gin.H{"error": "Key not found"})
		return
	}

	repository.DB.Model(&key).Updates(map[string]interface{}{
		"owner_name": b.OwnerName,
		"is_active":  b.IsActive,
	})

	repository.RDB.Del(context.Background(), "auth:"+key.KeyHash)

	c.JSON(200, key)
}

func HandleDeleteKey(c *gin.Context) {
	id := c.Param("id")
	var key models.APIKey
	if err := repository.DB.First(&key, id).Error; err == nil {
		repository.RDB.Del(context.Background(), "auth:"+key.KeyHash)

		if err := repository.DB.Model(&key).Association("Permissions").Clear(); err != nil {
			c.JSON(500, gin.H{"error": "Failed to clear permissions: " + err.Error()})
			return
		}

		if err := repository.DB.Delete(&key).Error; err != nil {
			c.JSON(500, gin.H{"error": "Failed to delete key: " + err.Error()})
			return
		}
	} else {
		c.JSON(404, gin.H{"error": "Key not found"})
		return
	}
	c.Status(204)
}

func HandleListUsers(c *gin.Context) {
	var users []models.User
	repository.DB.Preload("Role").Find(&users)
	c.JSON(200, users)
}

func HandleDeleteUser(c *gin.Context) {
	id := c.Param("id")
	var user models.User
	if err := repository.DB.First(&user, id).Error; err == nil {
		repository.InvalidateUserCache(user.ID)
		repository.DB.Delete(&user)
		c.Status(204)
		return
	}
	c.JSON(404, gin.H{"error": "User not found"})
}

func HandleListRoles(c *gin.Context) {
	var roles []models.Role
	repository.DB.Preload("Permissions").Find(&roles)
	c.JSON(200, roles)
}

func HandleUpdateRole(c *gin.Context) {
	id := c.Param("id")
	var b struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if err := c.BindJSON(&b); err != nil {
		c.Status(400)
		return
	}
	repository.DB.Model(&models.Role{}).Where("id = ?", id).Updates(models.Role{Name: b.Name, Description: b.Description})
	c.Status(200)
}

func HandleDeleteRole(c *gin.Context) {
	id := c.Param("id")
	var count int64
	repository.DB.Model(&models.User{}).Where("role_id = ?", id).Count(&count)

	if count > 0 {
		c.JSON(400, gin.H{"error": "Cannot delete role: users are still assigned to it"})
		return
	}

	repository.DB.Delete(&models.Role{}, id)
	c.Status(204)
}
