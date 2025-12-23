package main

import (
	"crypto/sha256"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var JWTSecret = []byte(os.Getenv("JWT_SECRET"))

func HashKey(key string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(key)))
}

func GenerateToken(id uint, role string) (string, error) {
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": id, "role": role, "exp": time.Now().Add(time.Hour * 24).Unix(),
	})
	return t.SignedString(JWTSecret)
}

func ValidateKey(key string) bool {
	h := HashKey(key)
	if v, _ := RDB.Get(ctx, "auth:"+h).Result(); v == "ok" { return true }
	
	var ak APIKey
	if err := DB.Where("key_hash = ? AND is_active = ?", h, true).First(&ak).Error; err == nil {
		RDB.Set(ctx, "auth:"+h, "ok", 15*time.Minute)
		return true
	}
	return false
}