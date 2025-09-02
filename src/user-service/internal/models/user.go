package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID              uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	Email           string    `gorm:"uniqueIndex;not null" json:"email"`
	Password        string    `gorm:"not null" json:"password"`
	Name            string    `gorm:"size:100;not null" json:"name"`
	CreatedAt       time.Time `gorm:"autoCreateTime" json:"createdAt"`
	LatestUpdatedAt time.Time `gorm:"autoUpdateTime" json:"latestUpdatedAt"`

	// Relations
	RoleID uuid.UUID `gorm:"type:uuid;not null" json:"roleID"`
	Role   Role      `gorm:"foreignKey:RoleID" json:"role"`
}

type Role struct {
	ID          uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	Name        string    `gorm:"uniqueIndex;not null" json:"name"`
	Description string    `gorm:"size:255" json:"description"`

	// One-to-many relation
	Users []User `gorm:"foreignKey:RoleID" json:"users"`
}

type RefreshToken struct {
	ID    uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	Token string    `gorm:"uniqueIndex;not null" json:"token"`
}
