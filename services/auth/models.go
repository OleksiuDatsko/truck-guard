package main

import "time"

type User struct {
	ID           uint   `gorm:"primaryKey"`
	Username     string `gorm:"unique;not null"`
	PasswordHash string `gorm:"not null"` // Зберігаємо bcrypt хеш
	Role         string `gorm:"default:operator"`
}

type APIKey struct {
	ID        uint      `gorm:"primaryKey"`
	KeyHash   string    `gorm:"unique;index;not null"` // Хеш ключа для пошуку
	OwnerName string    `json:"owner_name"`
	IsActive  bool      `gorm:"default:true"`
	CreatedAt time.Time
}