package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
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

	var role Role
	roleName := b.Role
	if roleName == "" {
		roleName = "operator"
	}
	DB.Where("name = ?", roleName).First(&role)

	u := User{Username: b.User, PasswordHash: string(h), RoleID: role.ID}
	if err := DB.Create(&u).Error; err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.Status(201)
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

	var u User
	if err := DB.Preload("Role.Permissions").Where("username = ?", b.User).First(&u).Error; err != nil {
		c.Status(401)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(b.Pass)); err != nil {
		c.Status(401)
		return
	}

	t, _ := GenerateToken(u)
	c.JSON(200, gin.H{"token": t})
}

func HandleValidate(c *gin.Context) {
	k := c.GetHeader("X-API-Key")
	if k != "" {
		if meta, valid := ValidateKeyAndGetMetadata(k); valid {
			c.Header("X-Camera-ID", meta.ID)
			c.Header("X-Camera-Name", meta.Name)
			c.Header("X-User-Permissions", strings.Join(meta.Permissions, ","))
			c.Status(200)
			return
		}
	}

	a := c.GetHeader("Authorization")
	if strings.HasPrefix(a, "Bearer ") {
		ts := strings.TrimPrefix(a, "Bearer ")
		token, err := jwt.Parse(ts, func(t *jwt.Token) (interface{}, error) { return JWTSecret, nil })

		if err == nil && token.Valid {
			if claims, ok := token.Claims.(jwt.MapClaims); ok {
				userID := uint(claims["user_id"].(float64))

				perms := GetUserPermissions(userID)

				c.Header("X-User-Permissions", strings.Join(perms, ","))
				c.Header("X-User-ID", fmt.Sprintf("%d", userID))

				c.Status(200)
				return
			}
		}
	}
	c.Status(401)
}

func HandleListPermissions(c *gin.Context) {
	var perms []Permission
	DB.Find(&perms)
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

	role := Role{Name: b.Name, Description: b.Description}
	if err := DB.Create(&role).Error; err != nil {
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

	var role Role
	if err := DB.First(&role, roleID).Error; err != nil {
		c.JSON(404, gin.H{"error": "Role not found"})
		return
	}

	var perms []Permission
	DB.Where("id IN ?", b.PermissionIDs).Find(&perms)

	if err := DB.Model(&role).Association("Permissions").Replace(perms); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
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

	DB.Model(&User{}).Where("id = ?", userIDStr).Update("role_id", b.RoleID)

	var id uint
	fmt.Sscanf(userIDStr, "%d", &id)
	InvalidateUserCache(id)

	c.Status(200)
}

func HandleListKeys(c *gin.Context) {
	var keys []APIKey
	DB.Preload("Permissions").Find(&keys)
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

	var perms []Permission
	DB.Where("id IN ?", b.PermissionIDs).Find(&perms)

	key := APIKey{
		KeyHash:     HashKey(rk),
		OwnerName:   b.Name,
		Permissions: perms,
	}

	if err := DB.Create(&key).Error; err != nil {
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

	var key APIKey
	if err := DB.First(&key, id).Error; err != nil {
		c.JSON(404, gin.H{"error": "Key not found"})
		return
	}

	DB.Model(&key).Update("is_active", b.IsActive)

	RDB.Del(ctx, "auth:"+key.KeyHash)

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

	var key APIKey
	if err := DB.First(&key, id).Error; err != nil {
		c.JSON(404, gin.H{"error": "Key not found"})
		return
	}

	var perms []Permission
	DB.Where("id IN ?", b.PermissionIDs).Find(&perms)

	DB.Model(&key).Association("Permissions").Replace(perms)

	RDB.Del(ctx, "auth:"+key.KeyHash)

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

	var key APIKey
	if err := DB.First(&key, id).Error; err != nil {
		c.JSON(404, gin.H{"error": "Key not found"})
		return
	}

	DB.Model(&key).Updates(map[string]interface{}{
		"owner_name": b.OwnerName,
		"is_active":  b.IsActive,
	})

	RDB.Del(ctx, "auth:"+key.KeyHash)

	c.JSON(200, key)
}

func HandleDeleteKey(c *gin.Context) {
	id := c.Param("id")
	var key APIKey
	if err := DB.First(&key, id).Error; err == nil {
		RDB.Del(ctx, "auth:"+key.KeyHash)
		DB.Delete(&key)
	}
	c.Status(204)
}

func HandleListUsers(c *gin.Context) {
	var users []User
	DB.Preload("Role").Find(&users)
	c.JSON(200, users)
}

func HandleDeleteUser(c *gin.Context) {
	id := c.Param("id")
	var user User
	if err := DB.First(&user, id).Error; err == nil {
		InvalidateUserCache(user.ID)
		DB.Delete(&user)
		c.Status(204)
		return
	}
	c.JSON(404, gin.H{"error": "User not found"})
}

func HandleListRoles(c *gin.Context) {
	var roles []Role
	DB.Preload("Permissions").Find(&roles)
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
	DB.Model(&Role{}).Where("id = ?", id).Updates(Role{Name: b.Name, Description: b.Description})
	c.Status(200)
}

func HandleDeleteRole(c *gin.Context) {
	id := c.Param("id")
	var count int64
	DB.Model(&User{}).Where("role_id = ?", id).Count(&count)

	if count > 0 {
		c.JSON(400, gin.H{"error": "Cannot delete role: users are still assigned to it"})
		return
	}

	DB.Delete(&Role{}, id)
	c.Status(204)
}
