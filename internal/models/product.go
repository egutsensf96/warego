package models

import "time"

type Product struct {
	Id_Product  uint `gorm:"primaryKey;autoIncrement:true"`
	Name        *string
	Description *string
	Cost        float32
	Stock       float32
	Image       *string
	Source      *string
	Category_Id Category `gorm:"foreignKey:Category_Id;references:Id_Category"`
	User_Id     User     `gorm:"foreignKey:User_Id;references:Id_User"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
