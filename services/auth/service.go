package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type CameraMetadata struct {
	ID   string
	Name string
}

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

func ValidateKeyAndGetMetadata(key string) (CameraMetadata, bool) {
	h := HashKey(key)
	
	if v, _ := RDB.Get(ctx, "auth:"+h).Result(); v != "" {
		var meta CameraMetadata
		json.Unmarshal([]byte(v), &meta)
		return meta, true
	}
	
	var ak APIKey
	if err := DB.Where("key_hash = ? AND is_active = ?", h, true).First(&ak).Error; err == nil {
		meta := CameraMetadata{
			ID:   fmt.Sprintf("%d", ak.ID),
			Name: ak.OwnerName,
		}
		val, _ := json.Marshal(meta)
		RDB.Set(ctx, "auth:"+h, val, 15*time.Minute)
		return meta, true
	}
	return CameraMetadata{}, false
}