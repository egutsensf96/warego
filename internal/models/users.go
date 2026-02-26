package models

import (
	"time"
)

type User struct {
	Id_User    uint `gorm:"primaryKey;autoIncrement:true"`
	Name       string
	LastName   string
	Cargo      string
	Permisos   string
	email      string
	password   string
	Company_Id Company `gorm:"foreignKey:Company_Id;references:Id_Company"`
	Role_Id    Role    `gorm:"foreignKey:Role_Id;references:Id_Role"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func (User) TableName() string {
	return "User"
}
