package main

import (
	"os"

	"github.com/gin-gonic/gin"
	"github.com/truckguard/ingestor/src/api/handlers"
	"github.com/truckguard/ingestor/src/api/middleware"
	"github.com/truckguard/ingestor/src/repository"
)

func main() {
	repository.InitRedis(os.Getenv("REDIS_ADDR"))

	endpoint := os.Getenv("MINIO_ENDPOINT")
	accessKey := os.Getenv("MINIO_ACCESS_KEY")
	secretKey := os.Getenv("MINIO_SECRET_KEY")
	repository.InitMinio(endpoint, accessKey, secretKey)

	r := gin.Default()
	r.POST("/ingest", middleware.RequirePermission("create:ingest"), handlers.HandleIngest)

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	r.Run(":" + port)
}
