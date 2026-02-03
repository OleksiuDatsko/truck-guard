package repository

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
	"github.com/truckguard/auth/src/models"
	"github.com/uptrace/opentelemetry-go-extra/otelgorm"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	DB        *gorm.DB
	RDB       *redis.Client
	ctx       = context.Background()
	JWTSecret = []byte(os.Getenv("JWT_SECRET"))
)

func InitDB(dsn string) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("Failed to connect to Auth Database")
	}

	if err := db.Use(otelgorm.NewPlugin()); err != nil {
		panic(err)
	}

	DB = db
	DB.AutoMigrate(&models.Permission{}, &models.Role{}, &models.User{}, &models.APIKey{})
}

func InitRedis(addr string) {
	RDB = redis.NewClient(&redis.Options{Addr: addr})
}

func HashKey(key string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(key)))
}

func GenerateToken(user models.User) (string, error) {
	claims := jwt.MapClaims{
		"user_id":  user.ID,
		"username": user.Username,
		"role":     user.Role.Name,
		"exp":      time.Now().Add(time.Hour * 24).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(JWTSecret)
}

func ValidateKeyAndGetMetadata(key string) (models.SourceMetadata, bool) {
	h := HashKey(key)

	if v, _ := RDB.Get(ctx, "auth:"+h).Result(); v != "" {
		var meta models.SourceMetadata
		json.Unmarshal([]byte(v), &meta)
		return meta, true
	}

	var ak models.APIKey
	if err := DB.Preload("Permissions").Where("key_hash = ? AND is_active = ?", h, true).First(&ak).Error; err == nil {
		var perms []string
		for _, p := range ak.Permissions {
			perms = append(perms, p.ID)
		}

		meta := models.SourceMetadata{
			ID:          fmt.Sprintf("%d", ak.ID),
			Name:        ak.OwnerName,
			Permissions: perms,
		}

		val, _ := json.Marshal(meta)
		RDB.Set(ctx, "auth:"+h, val, 15*time.Minute)
		return meta, true
	}

	return models.SourceMetadata{}, false
}

func GetUserPermissions(userID uint) []string {
	key := fmt.Sprintf("user_perms:%d", userID)

	val, _ := RDB.Get(ctx, key).Result()
	if val != "" {
		var perms []string
		json.Unmarshal([]byte(val), &perms)
		return perms
	}

	var u models.User
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
