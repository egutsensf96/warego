package models

import "time"

type Category struct {
	Id_Category uint `gorm:"primaryKey;autoIncrement:true"`
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
