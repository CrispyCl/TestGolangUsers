package models

import "time"

type UserRole string

const (
	UserRoleUnspecified UserRole = "UNSPECIFIED"
	UserRoleUser        UserRole = "USER"
	UserRoleManager     UserRole = "MANAGER"
	UserRoleAdmin       UserRole = "ADMIN"
)

type User struct {
	ID        int64     `json:"id"`
	Email     string    `json:"email"`
	PassHash  []byte    `json:"pass_hash"`
	Role      UserRole  `json:"role"`
	CreatedAt time.Time `json:"created_at"`
	LastSeen  time.Time `json:"last_seen"`
}
