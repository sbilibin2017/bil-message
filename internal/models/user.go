package models

import (
	"time"

	"github.com/google/uuid"
)

// UserDB представляет пользователя в бд.
type UserDB struct {
	UserUUID  uuid.UUID `json:"user_uuid" db:"user_uuid"`
	Username  string    `json:"username" db:"username"`
	Password  string    `json:"password" db:"password"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}
