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
	Email      string
	Password   string `-`
	Company_Id int
	Company    Company `gorm:"foreignKey:Company_Id;references:Id_Company"`
	Role_Id    int
	Role       Role `gorm:"foreignKey:Role_Id;references:Id_Role"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func (User) TableName() string {
	return "User"
}
