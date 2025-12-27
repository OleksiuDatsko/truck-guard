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
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Permissions []string `json:"permissions"`
}

func HashKey(key string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(key)))
}

func GenerateToken(user User) (string, error) {
	claims := jwt.MapClaims{
		"user_id":  user.ID,
		"username": user.Username,
		"role":     user.Role.Name,
		"exp":      time.Now().Add(time.Hour * 24).Unix(),
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
	if err := DB.Preload("Permissions").Where("key_hash = ? AND is_active = ?", h, true).First(&ak).Error; err == nil {
		var perms []string
		for _, p := range ak.Permissions {
			perms = append(perms, p.ID)
		}

		meta := CameraMetadata{
			ID:          fmt.Sprintf("%d", ak.ID),
			Name:        ak.OwnerName,
			Permissions: perms,
		}

		val, _ := json.Marshal(meta)
		RDB.Set(ctx, "auth:"+h, val, 15*time.Minute)
		return meta, true
	}

	return CameraMetadata{}, false
}

func GetUserPermissions(userID uint) []string {
	key := fmt.Sprintf("user_perms:%d", userID)

	val, _ := RDB.Get(ctx, key).Result()
	if val != "" {
		var perms []string
		json.Unmarshal([]byte(val), &perms)
		return perms
	}

	var u User
	DB.Preload("Role.Permissions").First(&u, userID)

	var perms []string
	for _, p := range u.Role.Permissions {
		perms = append(perms, p.ID)
	}

	valJSON, _ := json.Marshal(perms)
	RDB.Set(ctx, key, valJSON, time.Hour)

	return perms
}

func InvalidateUserCache(userID uint) {
	key := fmt.Sprintf("user_perms:%d", userID)
	RDB.Del(ctx, key)
}
