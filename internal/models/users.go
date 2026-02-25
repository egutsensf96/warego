package models

import "gorm.io/gorm"

type UserModel struct {
	Id_User uint `gorm:"primaryKey;autoIncrement:true"`
	Name   *string
	LastName *string
	Cargo *string
	Permisos *string
	email    string
	password string
	Company_Id CompanyModel `gorm:"foreignKey:Company_Id;references:Id_Company"`
	Role_Id RoleModel `gorm:"foreignKey:Role_Id;references:Id_Role"`
	CreatedAt    time.Time      
  	UpdatedAt    time.Time   

}
