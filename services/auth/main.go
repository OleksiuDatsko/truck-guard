package main

import (
	"os"

	"github.com/gin-gonic/gin"
	"github.com/truckguard/auth/src/api/handlers"
	"github.com/truckguard/auth/src/api/middleware"
	"github.com/truckguard/auth/src/repository"
)

func main() {
	repository.InitDB(os.Getenv("DATABASE_URL"))
	repository.InitRedis(os.Getenv("REDIS_ADDR"))

	seedData()

	r := gin.Default()

	r.POST("/login", handlers.HandleLogin)
	r.GET("/validate", handlers.HandleValidate)
	r.GET("/health", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })
	r.POST("/register", middleware.RequirePermission("create:users"), handlers.HandleRegister)

	admin := r.Group("/admin")
	{
		// Користувачі
		admin.GET("/users", middleware.RequirePermission("read:users"), handlers.HandleListUsers)
		admin.GET("/users/:id", middleware.RequirePermission("read:users"), handlers.HandleGetUser)
		admin.PUT("/users/:id/role", middleware.RequirePermission("update:users"), handlers.HandleUpdateUserRole)
		admin.DELETE("/users/:id", middleware.RequirePermission("delete:users"), handlers.HandleDeleteUser)

		// Ролі
		admin.GET("/roles", middleware.RequirePermission("read:roles"), handlers.HandleListRoles)
		admin.POST("/roles", middleware.RequirePermission("create:roles"), handlers.HandleCreateRole)
		admin.PUT("/roles/:id", middleware.RequirePermission("update:roles"), handlers.HandleUpdateRole)
		admin.DELETE("/roles/:id", middleware.RequirePermission("delete:roles"), handlers.HandleDeleteRole)
		admin.POST("/roles/:id/permissions", middleware.RequirePermission("update:roles"), handlers.HandleAssignPermissionsToRole)

		// Ключі (IoT)
		admin.GET("/keys", middleware.RequirePermission("read:keys"), handlers.HandleListKeys)
		admin.POST("/keys", middleware.RequirePermission("create:keys"), handlers.HandleCreateKeyWithPerms)
		admin.DELETE("/keys/:id", middleware.RequirePermission("delete:keys"), handlers.HandleDeleteKey)
		admin.PUT("/keys/:id/permissions", middleware.RequirePermission("update:keys"), handlers.HandleAssignPermissionsToKey)
		admin.PUT("/keys/:id", middleware.RequirePermission("update:keys"), handlers.HandleUpdateKey)

		admin.GET("/permissions", middleware.RequirePermission("read:roles"), handlers.HandleListPermissions)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	r.Run(":" + port)
}
