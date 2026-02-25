package models

type TrackerModel struct {
	Id_TrackerModel uint `gorm:"primaryKey;autoIncrement:true"`
	User_Id UserModel `gorm:"foreignKey:User_Id;references:Id_User"`
	Event *string
	Company_Id CompanyModel `gorm:"foreignKey:Company_Id;references:Id_Company"`
	UpdatedAt    time.Time 
}