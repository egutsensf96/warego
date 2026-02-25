package models

import "time"

type Role struct {
	Id_Role     uint `gorm:"primaryKey;autoIncrement:true"`
	description *string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
