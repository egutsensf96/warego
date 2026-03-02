package models

import "time"

type Role struct {
	Id_Role     uint      `gorm:"primaryKey;autoIncrement:true"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at,omitempty"`
	UpdatedAt   time.Time `json:"updated_at,omitempty"`
}

func (Role) TableName() string {
	return "Role"
}
