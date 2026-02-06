package main

import (
	"os"

	"github.com/truckguard/auth/src/models"
	"github.com/truckguard/auth/src/repository"
	"golang.org/x/crypto/bcrypt"
)

func seedData() {
	perms := []models.Permission{
		// Auth
		{ID: "read:users", Name: "Користувачі: Перегляд", Module: "auth"},
		{ID: "create:users", Name: "Користувачі: Створення", Module: "auth"},
		{ID: "update:users", Name: "Користувачі: Редагування", Module: "auth"},
		{ID: "delete:users", Name: "Користувачі: Видалення", Module: "auth"},
		{ID: "manage:users", Name: "Користувачі: Повний доступ", Module: "auth"},

		{ID: "read:roles", Name: "Ролі: Перегляд", Module: "auth"},
		{ID: "manage:roles", Name: "Ролі: Повний доступ", Module: "auth"},

		{ID: "read:keys", Name: "API Ключі: Перегляд", Module: "auth"},
		{ID: "manage:keys", Name: "API Ключі: Повний доступ", Module: "auth"},

		// Ingest
		{ID: "create:ingest", Name: "Імпорт: Створення", Module: "ingestor"},

		// Core
		{ID: "read:cameras", Name: "Камери: Перегляд", Module: "core"},
		{ID: "manage:cameras", Name: "Камери: Повний доступ", Module: "core"},

		{ID: "read:scales", Name: "Ваги: Перегляд", Module: "core"},
		{ID: "manage:scales", Name: "Ваги: Повний доступ", Module: "core"},

		{ID: "read:events", Name: "Події: Перегляд", Module: "core"},
		{ID: "create:events", Name: "Події: Створення", Module: "core"},
		{ID: "update:events", Name: "Події: Корекція", Module: "core"},

		{ID: "read:permits", Name: "Перепустки: Перегляд", Module: "core"},
		{ID: "create:permits", Name: "Перепустки: Створення", Module: "core"},
		{ID: "update:permits", Name: "Перепустки: Редагування", Module: "core"},

		{ID: "read:settings", Name: "Налаштування: Перегляд", Module: "core"},
		{ID: "update:settings", Name: "Налаштування: Зміна", Module: "core"},
		{ID: "manage:settings", Name: "Налаштування: Повний доступ", Module: "core"},

		{ID: "read:audit", Name: "Аудит: Перегляд", Module: "auth"},
	}

	for _, p := range perms {
		repository.DB.Save(&p)
	}

	// 1. Адмін - має все
	var adminRole models.Role
	repository.DB.FirstOrCreate(&adminRole, models.Role{Name: "admin", Description: "Повний доступ до системи"})
	repository.DB.Model(&adminRole).Association("Permissions").Replace(perms)

	// 2. Менеджер - може редагувати, але не видаляти (ієрархія 'update' включає 'read')
	var managerRole models.Role
	repository.DB.FirstOrCreate(&managerRole, models.Role{Name: "manager", Description: "Керівник (редагування без видалення)"})
	managerPermIDs := []string{
		"update:users", "update:permits", "update:events", "update:settings",
		"update:cameras", "update:scales", "read:roles", "read:keys",
	}
	managerPerms := []models.Permission{}
	repository.DB.Where("id IN ?", managerPermIDs).Find(&managerPerms)
	repository.DB.Model(&managerRole).Association("Permissions").Replace(managerPerms)

	// 3. Оператор - тільки перегляд та створення подій/перепусток
	var operatorRole models.Role
	repository.DB.FirstOrCreate(&operatorRole, models.Role{Name: "operator", Description: "Оператор (перегляд та створення)"})
	operatorPermIDs := []string{
		"read:permits", "create:permits", "read:events", "create:events", "read:cameras",
	}
	operatorPerms := []models.Permission{}
	repository.DB.Where("id IN ?", operatorPermIDs).Find(&operatorPerms)
	repository.DB.Model(&operatorRole).Association("Permissions").Replace(operatorPerms)

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
			println("Адміністратора за замовчуванням успішно створено: admin")
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
				OwnerName:   "Системний воркер",
				IsActive:    true,
				Permissions: workerPerms,
			}
			repository.DB.Create(&newKey)
			println("API ключ для системного воркера успішно додано")
		}
	}

	rules := []models.PolicyRule{
		// Auth Service
		{Method: "POST", PathPattern: `^/auth/register$`, RequiredPermission: "create:users", Description: "Реєстрація користувачів"},
		{Method: "GET", PathPattern: `^/auth/admin/users.*`, RequiredPermission: "read:users", Description: "Перегляд користувачів"},
		{Method: "PUT", PathPattern: `^/auth/admin/users/.*/role$`, RequiredPermission: "update:users", Description: "Зміна ролі користувача"},
		{Method: "DELETE", PathPattern: `^/auth/admin/users/.*`, RequiredPermission: "delete:users", Description: "Видалення користувача"},
		{Method: "GET", PathPattern: `^/auth/admin/roles.*`, RequiredPermission: "read:roles", Description: "Перегляд ролей"},
		{Method: "POST", PathPattern: `^/auth/admin/roles.*`, RequiredPermission: "manage:roles", Description: "Створення ролей"},
		{Method: "*", PathPattern: `^/auth/admin/roles/.*`, RequiredPermission: "manage:roles", Description: "Керування ролями"},
		{Method: "GET", PathPattern: `^/auth/admin/keys.*`, RequiredPermission: "read:keys", Description: "Перегляд ключів"},
		{Method: "POST", PathPattern: `^/auth/admin/keys.*`, RequiredPermission: "manage:keys", Description: "Створення ключів"},
		{Method: "*", PathPattern: `^/auth/admin/keys/.*`, RequiredPermission: "manage:keys", Description: "Керування ключами"},
		{Method: "GET", PathPattern: `^/auth/admin/permissions$`, RequiredPermission: "read:roles", Description: "Список всіх дозволів"},

		// Ingestor Service
		{Method: "POST", PathPattern: `^/ingest/.*`, RequiredPermission: "create:ingest", Description: "Імпорт даних"},

		// Core Service: CONFIGS
		{Method: "GET", PathPattern: `^/api/configs/cameras.*`, RequiredPermission: "read:cameras", Description: "Перегляд камер"},
		{Method: "POST", PathPattern: `^/api/configs/cameras.*`, RequiredPermission: "manage:cameras", Description: "Додавання камер"},
		{Method: "PUT", PathPattern: `^/api/configs/cameras/.*`, RequiredPermission: "manage:cameras", Description: "Редагування камер"},
		{Method: "DELETE", PathPattern: `^/api/configs/cameras/.*`, RequiredPermission: "manage:cameras", Description: "Видалення камер"},

		{Method: "GET", PathPattern: `^/api/configs/scales.*`, RequiredPermission: "read:scales", Description: "Перегляд ваг"},
		{Method: "POST", PathPattern: `^/api/configs/scales.*`, RequiredPermission: "manage:scales", Description: "Додавання ваг"},
		{Method: "PUT", PathPattern: `^/api/configs/scales/.*`, RequiredPermission: "manage:scales", Description: "Редагування ваг"},
		{Method: "DELETE", PathPattern: `^/api/configs/scales/.*`, RequiredPermission: "manage:scales", Description: "Видалення ваг"},

		{Method: "GET", PathPattern: `^/api/configs/settings.*`, RequiredPermission: "read:settings", Description: "Перегляд налаштувань"},
		{Method: "POST", PathPattern: `^/api/configs/settings.*`, RequiredPermission: "update:settings", Description: "Зміна налаштувань"},

		{Method: "GET", PathPattern: `^/api/configs/excluded-plates.*`, RequiredPermission: "read:settings", Description: "Перегляд чорного списку"},
		{Method: "POST", PathPattern: `^/api/configs/excluded-plates.*`, RequiredPermission: "update:settings", Description: "Додавання у чорний список"},
		{Method: "DELETE", PathPattern: `^/api/configs/excluded-plates/.*`, RequiredPermission: "update:settings", Description: "Видалення з чорного списку"},

		// Core Service: DATA
		{Method: "GET", PathPattern: `^/api/data/.*`, RequiredPermission: "read:settings", Description: "Перегляд довідників"},
		{Method: "*", PathPattern: `^/api/data/.*`, RequiredPermission: "manage:settings", Description: "Керування довідниками"},

		// Core Service: EVENTS
		{Method: "GET", PathPattern: `^/api/events/.*`, RequiredPermission: "read:events", Description: "Перегляд подій"},
		{Method: "POST", PathPattern: `^/api/events/.*`, RequiredPermission: "create:events", Description: "Реєстрація подій"},
		{Method: "PATCH", PathPattern: `^/api/events/plate/.*`, RequiredPermission: "update:events", Description: "Корекція номерів"},

		// Core Service: PERMITS
		{Method: "GET", PathPattern: `^/api/permits.*`, RequiredPermission: "read:permits", Description: "Перегляд перепусток"},
		{Method: "POST", PathPattern: `^/api/permits.*`, RequiredPermission: "create:permits", Description: "Створення перепусток"},
		{Method: "PUT", PathPattern: `^/api/permits/.*`, RequiredPermission: "update:permits", Description: "Оновлення перепусток"},

		// Core Service: USERS
		{Method: "GET", PathPattern: `^/api/users$`, RequiredPermission: "read:users", Description: "Список профілів"},
		{Method: "GET", PathPattern: `^/api/users/me`, RequiredPermission: "", Description: "Мій профіль"},
		{Method: "PUT", PathPattern: `^/api/users/me`, RequiredPermission: "", Description: "Оновлення мого профілю"},
		{Method: "POST", PathPattern: `^/api/users/$`, RequiredPermission: "create:users", Description: "Створення профілю"},
		{Method: "*", PathPattern: `^/api/users/.*`, RequiredPermission: "update:users", Description: "Керування профілями"},
	}

	for _, r := range rules {
		var existing models.PolicyRule
		if err := repository.DB.Where("method = ? AND path_pattern = ?", r.Method, r.PathPattern).First(&existing).Error; err != nil {
			repository.DB.Create(&r)
		}
	}
}
