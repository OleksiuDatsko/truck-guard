package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/truckguard/auth/src/models"
)

// CheckAccess перевіряє доступ, використовуючи регулярні вирази на рівні БД та кешування результатів у Redis.
func CheckAccess(method, path string) ([]string, bool, error) {
	cacheKey := fmt.Sprintf("auth:policy:%s:%s", strings.ToUpper(method), path)

	// 1. Спробуємо отримати з Redis
	if val, _ := RDB.Get(context.Background(), cacheKey).Result(); val != "" {
		if val == "UNMANAGED" {
			slog.Debug("Policy cache hit: unmanaged path", "path", path)
			return nil, false, nil // Не керується політиками
		}
		var reqPerms []string
		json.Unmarshal([]byte(val), &reqPerms)
		slog.Debug("Policy cache hit: rules found", "path", path, "required_perms", reqPerms)
		return reqPerms, true, nil // Керується, повертаємо список необхідних прав
	}

	slog.Info("Policy cache miss: querying database", "method", method, "path", path)

	// 2. Якщо в кеші немає — йдемо в Postgres
	// Спочатку перевіряємо, чи шлях взагалі керується хоч якимось правилом
	var allRulesForPath []models.PolicyRule
	if err := DB.Order("id asc").Where("? ~ path_pattern", path).Find(&allRulesForPath).Error; err != nil {
		return nil, false, err
	}

	if len(allRulesForPath) == 0 {
		// Шлях не керується ніякими правилами
		RDB.Set(context.Background(), cacheKey, "UNMANAGED", 15*time.Minute)
		return nil, false, nil
	}

	// 3. Фільтруємо правила за методом
	var requiredPerms []string
	for _, rule := range allRulesForPath {
		if rule.Method == "*" || strings.EqualFold(rule.Method, method) {
			requiredPerms = append(requiredPerms, rule.RequiredPermission)
		}
	}

	// 4. Кешуємо результат (навіть якщо список порожній — це означає "керований шлях, але метод не дозволений")
	val, _ := json.Marshal(requiredPerms)
	RDB.Set(context.Background(), cacheKey, string(val), 15*time.Minute)

	return requiredPerms, true, nil
}

// EvaluateAccess об’єднує перевірку політик та прав користувача
func EvaluateAccess(method, path string, userPerms []string) (bool, error) {
	reqPerms, isManaged, err := CheckAccess(method, path)
	if err != nil {
		return false, err
	}

	if !isManaged {
		return true, nil // Allow by default for unmanaged paths
	}

	// Якщо шлях керований, але жодне правило не підійшло під метод
	if len(reqPerms) == 0 {
		slog.Warn("Access denied: path is managed but method not allowed", "method", method, "path", path)
		return false, nil
	}
	// Перевіряємо, чи є у користувача хоча б одне з необхідних прав
	for _, rp := range reqPerms {
		if rp == "" || HasPermission(userPerms, rp) {
			return true, nil
		}
	}

	slog.Warn("Access denied: insufficient permissions", "method", method, "path", path, "required", reqPerms)
	return false, nil
}

// ValidatePermissions checks if all targetPerms are covered by userPerms (respecting hierarchy).
// This prevents permission escalation when assigning roles or keys.
func ValidatePermissions(userPerms []string, targetPerms []string) error {
	for _, tp := range targetPerms {
		if !HasPermission(userPerms, tp) {
			return fmt.Errorf("insufficient permission to grant: %s", tp)
		}
	}
	return nil
}

func HasPermission(userPerms []string, required string) bool {
	if required == "" {
		return true
	}

	for _, p := range userPerms {
		if p == "admin" || p == required {
			return true
		}

		// Підтримка формату action:resource (наприклад, update:permits)
		partsRequired := strings.Split(required, ":")
		partsUser := strings.Split(p, ":")

		if len(partsRequired) == 2 && len(partsUser) == 2 {
			actionUser := partsUser[0]
			resourceUser := partsUser[1]
			actionReq := partsRequired[0]
			resourceReq := partsRequired[1]

			// Ресурс має збігатися (або бути *)
			if resourceUser == "*" || resourceUser == resourceReq {
				// Ієрархія дій:
				// manage > delete > update > create > read
				switch actionUser {
				case "manage", "admin":
					return true
				case "delete":
					if actionReq == "delete" || actionReq == "update" || actionReq == "create" || actionReq == "read" {
						return true
					}
				case "update":
					if actionReq == "update" || actionReq == "create" || actionReq == "read" {
						return true
					}
				case "create":
					if actionReq == "create" || actionReq == "read" {
						return true
					}
				case "read":
					if actionReq == "read" {
						return true
					}
				}
			}
		}
	}
	return false
}
