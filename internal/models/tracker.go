package models

import "time"

type Tracker struct {
	Id_TrackerModel uint `gorm:"primaryKey;autoIncrement:true"`
	User_Id         int
	User            User `gorm:"foreignKey:User_Id;references:Id_User"`
	Event           string
	Company_Id      int
	Company         Company `gorm:"foreignKey:Company_Id;references:Id_Company"`
	UpdatedAt       time.Time
}

func (Tracker) TableName() string {
	return "Tracker"
}
