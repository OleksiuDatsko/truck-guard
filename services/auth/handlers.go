package main

import (
	"crypto/rand"
	"encoding/hex"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"strings"
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
			c.Status(200)
			return
		}
	}

	a := c.GetHeader("Authorization")
	if strings.HasPrefix(a, "Bearer ") {
		ts := strings.TrimPrefix(a, "Bearer ")
		token, _ := jwt.Parse(ts, func(t *jwt.Token) (interface{}, error) { return JWTSecret, nil })
		if token != nil && token.Valid {
			c.Status(200)
			return
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
	userID := c.Param("id")
	var b struct {
		RoleID uint `json:"role_id"`
	}
	if err := c.BindJSON(&b); err != nil {
		c.Status(400)
		return
	}
	DB.Model(&User{}).Where("id = ?", userID).Update("role_id", b.RoleID)
	c.Status(200)
}

func HandleCreateKey(c *gin.Context) {
	var b struct{ Name string }
	if err := c.BindJSON(&b); err != nil {
		c.Status(400)
		return
	}
	rb := make([]byte, 16)
	rand.Read(rb)
	rk := hex.EncodeToString(rb)
	DB.Create(&APIKey{KeyHash: HashKey(rk), OwnerName: b.Name})
	c.JSON(201, gin.H{"api_key": rk})
}
