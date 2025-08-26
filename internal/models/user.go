package models

import (
	"time"

	"github.com/google/uuid"
)

type UserDB struct {
	UserUUID  uuid.UUID `db:"user_uuid"`
	Username  string    `db:"username"`
	Password  string    `db:"password"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}
