package dtos

import "time"

type AuthRole struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type AuthUser struct {
	ID        uint       `json:"id"`
	Username  string     `json:"username"`
	RoleID    uint       `json:"role_id"`
	Role      AuthRole   `json:"role"`
	CreatedAt time.Time  `json:"created_at"`
	LastLogin *time.Time `json:"last_login"`
}
