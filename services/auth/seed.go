package main

import (
	"os"

	"github.com/truckguard/auth/src/models"
	"github.com/truckguard/auth/src/repository"
	"golang.org/x/crypto/bcrypt"
)

func seedData() {
	perms := []models.Permission{
		// Модуль: Авторизація
		{ID: "read:users", Name: "Перегляд користувачів", Description: "Дозволяє переглядати список користувачів та їх деталі", Module: "auth"},
		{ID: "create:users", Name: "Створення користувачів", Description: "Дозволяє створювати нових користувачів", Module: "auth"},
		{ID: "update:users", Name: "Оновлення користувачів", Description: "Дозволяє редагувати дані користувачів", Module: "auth"},
		{ID: "delete:users", Name: "Видалення користувачів", Description: "Дозволяє видаляти користувачів", Module: "auth"},
		{ID: "read:roles", Name: "Перегляд ролей", Description: "Дозволяє переглядати список ролей", Module: "auth"},
		{ID: "create:roles", Name: "Створення ролей", Description: "Дозволяє створювати нові ролі", Module: "auth"},
		{ID: "update:roles", Name: "Оновлення ролей", Description: "Дозволяє редагувати існуючі ролі", Module: "auth"},
		{ID: "delete:roles", Name: "Видалення ролей", Description: "Дозволяє видаляти ролі", Module: "auth"},
		{ID: "manage:settings", Name: "Керування налаштуваннями", Description: "Доступ до налаштувань системи", Module: "auth"},
		{ID: "view:audit", Name: "Перегляд аудиту", Description: "Доступ до журналу аудиту", Module: "auth"},
		{ID: "auth:login", Name: "Доступ до входу", Description: "Дозвіл на вхід в систему", Module: "auth"},
		{ID: "self:profile", Name: "Власний профіль", Description: "Доступ до власного профілю", Module: "auth"},
		{ID: "read:keys", Name: "Перегляд API ключів", Description: "Дозволяє переглядати API ключі", Module: "auth"},
		{ID: "create:keys", Name: "Створення API ключів", Description: "Дозволяє створювати API ключі", Module: "auth"},
		{ID: "update:keys", Name: "Оновлення API ключів", Description: "Дозволяє редагувати API ключі", Module: "auth"},
		{ID: "delete:keys", Name: "Видалення API ключів", Description: "Дозволяє видаляти API ключі", Module: "auth"},

		// Модуль: Інджектор
		{ID: "create:ingest", Name: "Створення даних імпорту", Description: "Дозволяє імпортувати дані з зовнішніх джерел", Module: "ingestor"},

		// Модуль: Ядро (Core)
		{ID: "read:trips", Name: "Перегляд поїздок", Description: "Дозволяє переглядати історію поїздок", Module: "core"},
		{ID: "create:events", Name: "Створення подій", Description: "Дозволяє створювати події (наприклад, розпізнавання номерів)", Module: "core"},
		{ID: "update:events", Name: "Оновлення подій", Description: "Дозволяє редагувати події", Module: "core"},
		{ID: "read:events", Name: "Перегляд подій", Description: "Дозволяє переглядати список подій", Module: "core"},
		{ID: "read:cameras", Name: "Перегляд камер", Description: "Дозволяє переглядати список камер", Module: "core"},
		{ID: "create:cameras", Name: "Створення камер", Description: "Дозволяє додавати нові камери", Module: "core"},
		{ID: "update:cameras", Name: "Оновлення камер", Description: "Дозволяє редагувати налаштування камер", Module: "core"},
		{ID: "delete:cameras", Name: "Видалення камер", Description: "Дозволяє видаляти камери", Module: "core"},
		{ID: "manage:configs", Name: "Керування конфігураціями", Description: "Доступ до конфігурацій ядра", Module: "core"},
		{ID: "read:presets", Name: "Перегляд пресетів", Description: "Дозволяє переглядати пресети конфігурацій", Module: "core"},
		{ID: "create:presets", Name: "Створення пресетів", Description: "Дозволяє створювати пресети", Module: "core"},
		{ID: "update:presets", Name: "Оновлення пресетів", Description: "Дозволяє оновлювати пресети", Module: "core"},
		{ID: "delete:presets", Name: "Видалення пресетів", Description: "Дозволяє видаляти пресети", Module: "core"},
		{ID: "read:scales", Name: "Перегляд ваг", Description: "Дозволяє переглядати список ваг", Module: "core"},
		{ID: "create:scales", Name: "Створення ваг", Description: "Дозволяє додавати ваги", Module: "core"},
		{ID: "update:scales", Name: "Оновлення ваг", Description: "Дозволяє редагувати ваги", Module: "core"},
		{ID: "delete:scales", Name: "Видалення ваг", Description: "Дозволяє видаляти ваги", Module: "core"},
		{ID: "read:gates", Name: "Перегляд воріт", Description: "Дозволяє переглядати список воріт", Module: "core"},
		{ID: "create:gates", Name: "Створення воріт", Description: "Дозволяє створювати ворота", Module: "core"},
		{ID: "update:gates", Name: "Оновлення воріт", Description: "Дозволяє редагувати ворота", Module: "core"},
		{ID: "delete:gates", Name: "Видалення воріт", Description: "Дозволяє видаляти ворота", Module: "core"},
		{ID: "read:settings", Name: "Перегляд налаштувань", Description: "Дозволяє переглядати системні налаштування", Module: "core"},
		{ID: "update:settings", Name: "Оновлення налаштувань", Description: "Дозволяє змінювати системні налаштування", Module: "core"},
		{ID: "read:excluded_plates", Name: "Перегляд виключених номерів", Description: "Дозволяє переглядати чорний список номерів", Module: "core"},
		{ID: "create:excluded_plates", Name: "Створення виключених номерів", Description: "Дозволяє додавати номери до чорного списку", Module: "core"},
		{ID: "delete:excluded_plates", Name: "Видалення виключених номерів", Description: "Дозволяє видаляти номери з чорного списку", Module: "core"},
		{ID: "read:permits", Name: "Перегляд перепусток", Description: "Дозволяє переглядати перепустки (для оператора - тільки по своєму посту)", Module: "core"},
		{ID: "read:permits:all", Name: "Перегляд всіх перепусток", Description: "Дозволяє переглядати всі перепустки незалежно від посту", Module: "core"},
		{ID: "read:flows", Name: "Перегляд потоків", Description: "Дозволяє переглядати налаштовані потоки", Module: "core"},
		{ID: "create:flows", Name: "Створення потоків", Description: "Дозволяє створювати нові потоки", Module: "core"},
		{ID: "update:flows", Name: "Оновлення потоків", Description: "Дозволяє редагувати потоки", Module: "core"},
		{ID: "delete:flows", Name: "Видалення потоків", Description: "Дозволяє видаляти потоки", Module: "core"},
	}

	for _, p := range perms {
		repository.DB.Save(&p)
	}

	var adminRole models.Role
	repository.DB.FirstOrCreate(&adminRole, models.Role{Name: "admin", Description: "Повний доступ"})
	repository.DB.Model(&adminRole).Association("Permissions").Replace(perms)

	var managerRole models.Role
	repository.DB.FirstOrCreate(&managerRole, models.Role{Name: "manager", Description: "Керуючий"})
	// Grant manager permissions (all read + specific manage)
	managerPerms := []models.Permission{}
	permIDs := []string{
		"read:users", "read:roles", "read:keys", "view:audit", "self:profile",
		"read:trips", "read:events", "read:cameras", "read:scales", "read:gates", "read:settings",
		"read:excluded_plates", "create:excluded_plates", "delete:excluded_plates",
		"read:permits:all", "read:permits", "read:flows",
	}
	repository.DB.Where("id IN ?", permIDs).Find(&managerPerms)
	repository.DB.Model(&managerRole).Association("Permissions").Replace(managerPerms)

	var operatorRole models.Role
	repository.DB.FirstOrCreate(&operatorRole, models.Role{Name: "operator", Description: "Стандартний доступ"})

	// Refine operator permissions
	operatorPermIDs := []string{
		"self:profile", "auth:login",
		"read:permits", // Scope restricted
		"read:events",
		"read:gates",
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
}
