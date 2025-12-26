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
		admin.PUT("/users/:id/role", HandleUpdateUserRole)

		admin.POST("/keys", HandleCreateKey)

		admin.GET("/permissions", HandleListPermissions)
		admin.POST("/roles", HandleCreateRole)
		admin.POST("/roles/:id/permissions", HandleAssignPermissionsToRole)
	}

	r.Run(":8080")
}

func seedData() {
	perms := []Permission{
		{ID: "read:vehicles", Name: "Read Vehicles", Module: "core"},
		{ID: "create:vehicles", Name: "Create Vehicles", Module: "core"},
		{ID: "update:vehicles", Name: "Update Vehicles", Module: "core"},
		{ID: "delete:vehicles", Name: "Delete Vehicles", Module: "core"},
		{ID: "read:reports", Name: "Read Reports", Module: "core"},
		{ID: "create:reports", Name: "Create Reports", Module: "core"},
		{ID: "update:reports", Name: "Update Reports", Module: "core"},
		{ID: "delete:reports", Name: "Delete Reports", Module: "core"},
		{ID: "read:users", Name: "Read Users", Module: "auth"},
		{ID: "create:users", Name: "Create Users", Module: "auth"},
		{ID: "update:users", Name: "Update Users", Module: "auth"},
		{ID: "delete:users", Name: "Delete Users", Module: "auth"},
		{ID: "read:roles", Name: "Read Roles", Module: "auth"},
		{ID: "create:roles", Name: "Create Roles", Module: "auth"},
		{ID: "update:roles", Name: "Update Roles", Module: "auth"},
		{ID: "delete:roles", Name: "Delete Roles", Module: "auth"},
		{ID: "export:data", Name: "Export Data", Module: "core"},
		{ID: "manage:settings", Name: "Manage Settings", Module: "auth"},
		{ID: "view:audit", Name: "View Audit", Module: "auth"},
		{ID: "auth:login", Name: "Login Access", Module: "auth"},
		{ID: "self:profile", Name: "Self Profile", Module: "auth"},
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
