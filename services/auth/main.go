package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/truckguard/auth/src/api/handlers"
	"github.com/truckguard/auth/src/api/middleware"
	"github.com/truckguard/auth/src/pkg/telemetry"
	"github.com/truckguard/auth/src/repository"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

func main() {
	logger := telemetry.NewLogger("truckguard-auth")
	slog.SetDefault(logger)

	if err := telemetry.Init("truckguard-auth"); err != nil {
		logger.Error("otel init failed", "error", err)
		os.Exit(1)
	}
	defer telemetry.Shutdown(context.Background())

	repository.InitDB(os.Getenv("DATABASE_URL"))
	repository.InitRedis(os.Getenv("REDIS_ADDR"))

	seedData()

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.Logger())
	r.Use(otelgin.Middleware("truckguard-auth"))

	r.POST("/login", handlers.HandleLogin)
	r.GET("/validate", handlers.HandleValidate)
	r.GET("/health", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })
	r.POST("/register", handlers.HandleRegister)

	admin := r.Group("/admin")
	{
		// Користувачі
		admin.GET("/users", handlers.HandleListUsers)
		admin.GET("/users/:id", handlers.HandleGetUser)
		admin.PUT("/users/:id/role", handlers.HandleUpdateUserRole)
		admin.DELETE("/users/:id", handlers.HandleDeleteUser)

		// Ролі
		admin.GET("/roles", handlers.HandleListRoles)
		admin.POST("/roles", handlers.HandleCreateRole)
		admin.PUT("/roles/:id", handlers.HandleUpdateRole)
		admin.DELETE("/roles/:id", handlers.HandleDeleteRole)
		admin.POST("/roles/:id/permissions", handlers.HandleAssignPermissionsToRole)

		// Ключі (IoT)
		admin.GET("/keys", handlers.HandleListKeys)
		admin.POST("/keys", handlers.HandleCreateKeyWithPerms)
		admin.DELETE("/keys/:id", handlers.HandleDeleteKey)
		admin.PUT("/keys/:id/permissions", handlers.HandleAssignPermissionsToKey)
		admin.PUT("/keys/:id", handlers.HandleUpdateKey)

		admin.GET("/permissions", handlers.HandleListPermissions)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	r.Run(":" + port)
}
