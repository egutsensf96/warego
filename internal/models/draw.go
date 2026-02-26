package models

import "time"

type Draw struct {
	Id_Draw    uint    `gorm:"primaryKey;autoIncrement:true"`
	Product_Id Product `gorm:"foreignKey:Product_Id;references:Id_Product"`
	Stock      float32
	Company_Id Company `gorm:"foreignKey:Company_Id;references:Id_Company"`
	User_Id    User    `gorm:"foreignKey:User_Id;references:Id_User"`
	CreatedAt  time.Time
}

func (Draw) TableName() string {
	return "Draw"
}
