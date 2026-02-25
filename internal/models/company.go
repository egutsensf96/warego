package models

import "time"

type Company struct {
	Id_Company  uint `gorm:"primaryKey;autoIncrement:true"`
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
