package main

import (
	"context"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	DB  *gorm.DB
	RDB *redis.Client
	ctx = context.Background()
)

func main() {
	db, _ := gorm.Open(postgres.Open(os.Getenv("DATABASE_URL")), &gorm.Config{})
	DB = db
	DB.AutoMigrate(&User{}, &APIKey{})

	RDB = redis.NewClient(&redis.Options{Addr: os.Getenv("REDIS_ADDR")})

	r := gin.Default()

	r.POST("/register", HandleRegister)
	r.POST("/keys", HandleCreateKey)
	r.POST("/login", HandleLogin)
	r.GET("/validate", HandleValidate)

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	r.Run(":8080")
}
