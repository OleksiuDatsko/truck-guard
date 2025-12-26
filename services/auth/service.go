package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var JWTSecret = []byte(os.Getenv("JWT_SECRET"))

type CameraMetadata struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func HashKey(key string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(key)))
}

func GenerateToken(user User) (string, error) {
	var permissions []string
	for _, p := range user.Role.Permissions {
		permissions = append(permissions, p.ID)
	}

	claims := jwt.MapClaims{
		"user_id":     user.ID,
		"username":    user.Username,
		"role":        user.Role.Name,
		"permissions": permissions,
		"exp":         time.Now().Add(time.Hour * 24).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(JWTSecret)
}

func ValidateKeyAndGetMetadata(key string) (CameraMetadata, bool) {
	h := HashKey(key)
	if v, _ := RDB.Get(ctx, "auth:"+h).Result(); v != "" {
		var meta CameraMetadata
		json.Unmarshal([]byte(v), &meta)
		return meta, true
	}

	var ak APIKey
	if err := DB.Where("key_hash = ? AND is_active = ?", h, true).First(&ak).Error; err == nil {
		meta := CameraMetadata{ID: fmt.Sprintf("%d", ak.ID), Name: ak.OwnerName}
		val, _ := json.Marshal(meta)
		RDB.Set(ctx, "auth:"+h, val, 15*time.Minute)
		return meta, true
	}
	return CameraMetadata{}, false
}
