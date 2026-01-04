package models

import "time"

type Permission struct {
	ID          string `gorm:"primaryKey" json:"id"`
	Name        string `gorm:"not null" json:"name"`
	Description string `json:"description"`
	Module      string `json:"module"`
}

type Role struct {
	ID          uint         `gorm:"primaryKey" json:"id"`
	Name        string       `gorm:"unique;not null" json:"name"`
	Description string       `json:"description"`
	Permissions []Permission `gorm:"many2many:role_permissions;" json:"permissions"`
}

type User struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Username     string    `gorm:"unique;not null" json:"username"`
	PasswordHash string    `gorm:"not null" json:"-"`
	RoleID       uint      `json:"role_id"`
	Role         Role      `gorm:"foreignKey:RoleID" json:"role"`
	CreatedAt    time.Time `json:"created_at"`
}

type APIKey struct {
	ID          uint         `gorm:"primaryKey" json:"id"`
	KeyHash     string       `gorm:"unique;index;not null" json:"-"`
	OwnerName   string       `json:"owner_name"`
	IsActive    bool         `gorm:"default:true" json:"is_active"`
	Permissions []Permission `gorm:"many2many:apikey_permissions;" json:"permissions"`
	CreatedAt   time.Time    `json:"created_at"`
}

type CameraMetadata struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Permissions []string `json:"permissions"`
}
