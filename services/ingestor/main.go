package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/truckguard/ingestor/src/api/handlers"
	"github.com/truckguard/ingestor/src/api/middleware"
	"github.com/truckguard/ingestor/src/pkg/telemetry"
	"github.com/truckguard/ingestor/src/repository"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

func main() {
	logger := telemetry.NewLogger("truckguard-ingestor")
	slog.SetDefault(logger)

	if err := telemetry.Init("truckguard-ingestor"); err != nil {
		logger.Error("otel init failed", "error", err)
		os.Exit(1)
	}
	defer telemetry.Shutdown(context.Background())

	repository.InitRedis(os.Getenv("REDIS_ADDR"))

	endpoint := os.Getenv("MINIO_ENDPOINT")
	accessKey := os.Getenv("MINIO_ACCESS_KEY")
	secretKey := os.Getenv("MINIO_SECRET_KEY")
	repository.InitMinio(endpoint, accessKey, secretKey)

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.Logger())
	r.Use(otelgin.Middleware("truckguard-ingestor"))

	ingestLines := r.Group("/ingest", middleware.RequirePermission("create:ingest"))
	{
		ingestLines.POST("/camera", handlers.HandleCameraIngest)
		ingestLines.POST("/weight", handlers.HandleWeightIngest)
	}

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	r.Run(":" + port)
}
