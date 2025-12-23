package main

import (
	"crypto/rand"
	"encoding/hex"
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
	u := User{Username: b.User, PasswordHash: string(h), Role: b.Role}

	if err := DB.Create(&u).Error; err != nil {
		if strings.Contains(err.Error(), "23505") {
			c.JSON(409, gin.H{"error": "username already taken"})
			return
		}
		c.JSON(500, gin.H{"e": err.Error()})
		return
	}
	c.Status(201)
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

	if err := DB.Create(&APIKey{KeyHash: HashKey(rk), OwnerName: b.Name}).Error; err != nil {
		c.Status(500)
		return
	}
	c.JSON(201, gin.H{"api_key": rk})
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
	if err := DB.Where("username = ?", b.User).First(&u).Error; err != nil {
		c.Status(401)
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(b.Pass)); err != nil {
		c.Status(401)
		return
	}

	t, _ := GenerateToken(u.ID, u.Role)
	c.JSON(200, gin.H{"token": t})
}

func HandleValidate(c *gin.Context) {
	k := c.GetHeader("X-API-Key")
	if k != "" && ValidateKey(k) {
		c.Status(200)
		return
	}

	a := c.GetHeader("Authorization")
	if strings.HasPrefix(a, "Bearer ") {
		ts := strings.TrimPrefix(a, "Bearer ")
		t, _ := jwt.Parse(ts, func(t *jwt.Token) (interface{}, error) { return JWTSecret, nil })
		if t != nil && t.Valid {
			c.Status(200)
			return
		}
	}
	c.Status(401)
}
