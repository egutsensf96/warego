package models

import "time"

type Tracker struct {
	Id_TrackerModel uint `gorm:"primaryKey;autoIncrement:true"`
	User_Id         User `gorm:"foreignKey:User_Id;references:Id_User"`
	Event           string
	Company_Id      Company `gorm:"foreignKey:Company_Id;references:Id_Company"`
	UpdatedAt       time.Time
}
