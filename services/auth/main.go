package main

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"os"
)

var (
	DB  *gorm.DB
	RDB *redis.Client
	ctx = context.Background()
)

func main() {
	db, _ := gorm.Open(postgres.Open(os.Getenv("DATABASE_URL")), &gorm.Config{})
	DB = db
	DB.AutoMigrate(&Permission{}, &Role{}, &User{}, &APIKey{})

	RDB = redis.NewClient(&redis.Options{Addr: os.Getenv("REDIS_ADDR")})

	seedData()

	r := gin.Default()

	r.POST("/login", HandleLogin)
	r.GET("/validate", HandleValidate)
	r.GET("/health", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })
	r.POST("/register", RequirePermission("create:users"), HandleRegister)

	admin := r.Group("/admin")
	admin.Use(RequirePermission("manage:settings"))
	{
		// Користувачі
		admin.GET("/users", RequirePermission("read:users"), HandleListUsers)
		admin.PUT("/users/:id/role", RequirePermission("update:users"), HandleUpdateUserRole)
		admin.DELETE("/users/:id", RequirePermission("delete:users"), HandleDeleteUser)

		// Ролі
		admin.GET("/roles", RequirePermission("read:roles"), HandleListRoles)
		admin.POST("/roles", RequirePermission("create:roles"), HandleCreateRole)
		admin.PUT("/roles/:id", RequirePermission("update:roles"), HandleUpdateRole)
		admin.DELETE("/roles/:id", RequirePermission("delete:roles"), HandleDeleteRole)
		admin.POST("/roles/:id/permissions", RequirePermission("update:roles"), HandleAssignPermissionsToRole)

		// Ключі (IoT)
		admin.GET("/keys", RequirePermission("read:keys"), HandleListKeys)
		admin.POST("/keys", RequirePermission("create:keys"), HandleCreateKeyWithPerms)
		admin.DELETE("/keys/:id", RequirePermission("delete:keys"), HandleDeleteKey)
		admin.PUT("/keys/:id/permissions", RequirePermission("update:keys"), HandleAssignPermissionsToKey)
		admin.PUT("/keys/:id", RequirePermission("update:keys"), HandleUpdateKey)

		admin.GET("/permissions", RequirePermission("read:roles"), HandleListPermissions)
	}

	r.Run(":8080")
}

func seedData() {
	perms := []Permission{
		{ID: "read:users", Name: "Read Users", Module: "auth"},
		{ID: "create:users", Name: "Create Users", Module: "auth"},
		{ID: "update:users", Name: "Update Users", Module: "auth"},
		{ID: "delete:users", Name: "Delete Users", Module: "auth"},
		{ID: "read:roles", Name: "Read Roles", Module: "auth"},
		{ID: "create:roles", Name: "Create Roles", Module: "auth"},
		{ID: "update:roles", Name: "Update Roles", Module: "auth"},
		{ID: "delete:roles", Name: "Delete Roles", Module: "auth"},
		{ID: "manage:settings", Name: "Manage Settings", Module: "auth"},
		{ID: "view:audit", Name: "View Audit", Module: "auth"},
		{ID: "auth:login", Name: "Login Access", Module: "auth"},
		{ID: "self:profile", Name: "Self Profile", Module: "auth"},
		{ID: "read:keys", Name: "Read API Keys", Module: "auth"},
		{ID: "create:keys", Name: "Create API Keys", Module: "auth"},
		{ID: "update:keys", Name: "Update API Keys", Module: "auth"},
		{ID: "delete:keys", Name: "Delete API Keys", Module: "auth"},
		{ID: "create:ingest", Name: "Create Ingestion Data", Module: "ingestor"},
	}

	for _, p := range perms {
		DB.FirstOrCreate(&p, Permission{ID: p.ID})
	}

	var adminRole Role
	DB.FirstOrCreate(&adminRole, Role{Name: "admin", Description: "Full Access"})
	DB.Model(&adminRole).Association("Permissions").Replace(perms)

	var operatorRole Role
	DB.FirstOrCreate(&operatorRole, Role{Name: "operator", Description: "Standard Access"})
}
