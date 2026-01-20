package main

import (
	"os"

	"github.com/truckguard/auth/src/models"
	"github.com/truckguard/auth/src/repository"
	"golang.org/x/crypto/bcrypt"
)

func seedData() {
	perms := []models.Permission{
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

		{ID: "read:trips", Name: "Read Trips", Module: "core"},
		{ID: "create:events", Name: "Create Events", Module: "core"},
		{ID: "update:events", Name: "Update Events", Module: "core"},
		{ID: "read:events", Name: "Read Events", Module: "core"},
		{ID: "read:cameras", Name: "Read Cameras", Module: "core"},
		{ID: "create:cameras", Name: "Create Cameras", Module: "core"},
		{ID: "update:cameras", Name: "Update Cameras", Module: "core"},
		{ID: "delete:cameras", Name: "Delete Cameras", Module: "core"},
		{ID: "manage:configs", Name: "Manage Configurations", Module: "core"},
		{ID: "read:presets", Name: "Read Presets", Module: "core"},
		{ID: "create:presets", Name: "Create Presets", Module: "core"},
		{ID: "update:presets", Name: "Update Presets", Module: "core"},
		{ID: "delete:presets", Name: "Delete Presets", Module: "core"},
		{ID: "read:scales", Name: "Read Scales", Module: "core"},
		{ID: "create:scales", Name: "Create Scales", Module: "core"},
		{ID: "update:scales", Name: "Update Scales", Module: "core"},
		{ID: "delete:scales", Name: "Delete Scales", Module: "core"},
		{ID: "read:gates", Name: "Read Gates", Module: "core"},
		{ID: "create:gates", Name: "Create Gates", Module: "core"},
		{ID: "update:gates", Name: "Update Gates", Module: "core"},
		{ID: "delete:gates", Name: "Delete Gates", Module: "core"},
		{ID: "read:settings", Name: "Read Settings", Module: "core"},
		{ID: "update:settings", Name: "Update Settings", Module: "core"},
		{ID: "read:excluded_plates", Name: "Read Excluded Plates", Module: "core"},
		{ID: "create:excluded_plates", Name: "Create Excluded Plates", Module: "core"},
		{ID: "delete:excluded_plates", Name: "Delete Excluded Plates", Module: "core"},
		{ID: "read:permits", Name: "Read Permits", Module: "core"},
		{ID: "read:flows", Name: "Read Flows", Module: "core"},
		{ID: "create:flows", Name: "Create Flows", Module: "core"},
		{ID: "update:flows", Name: "Update Flows", Module: "core"},
		{ID: "delete:flows", Name: "Delete Flows", Module: "core"},
	}

	for _, p := range perms {
		repository.DB.FirstOrCreate(&p, models.Permission{ID: p.ID})
	}

	var adminRole models.Role
	repository.DB.FirstOrCreate(&adminRole, models.Role{Name: "admin", Description: "Full Access"})
	repository.DB.Model(&adminRole).Association("Permissions").Replace(perms)

	var operatorRole models.Role
	repository.DB.FirstOrCreate(&operatorRole, models.Role{Name: "operator", Description: "Standard Access"})

	adminUsername := "admin"
	adminPassword := os.Getenv("ADMIN_DEFAULT_PASSWORD")
	if adminPassword == "" {
		adminPassword = "admin123"
	}
	var adminUser models.User
	err := repository.DB.Where("username = ?", adminUsername).First(&adminUser).Error
	if err != nil {
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(adminPassword), bcrypt.DefaultCost)

		newAdmin := models.User{
			Username:     adminUsername,
			PasswordHash: string(hashedPassword),
			RoleID:       adminRole.ID,
			Role:         adminRole,
		}

		if createErr := repository.DB.Create(&newAdmin).Error; createErr == nil {
			println("Successfully created default admin user: admin")
		}
	}

	workerKey := os.Getenv("WORKER_SYSTEM_KEY")
	if workerKey != "" {
		h := repository.HashKey(workerKey)
		var existingKey models.APIKey
		err := repository.DB.Where("key_hash = ?", h).First(&existingKey).Error
		if err != nil {
			workerPerms := []models.Permission{}
			repository.DB.Where("id IN ?", []string{"manage:configs", "create:events", "read:trips"}).Find(&workerPerms)

			newKey := models.APIKey{
				KeyHash:     h,
				OwnerName:   "System Worker",
				IsActive:    true,
				Permissions: workerPerms,
			}
			repository.DB.Create(&newKey)
			println("Successfully seeded System Worker API Key")
		}
	}
}
