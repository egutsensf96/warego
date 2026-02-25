package models

type RoleModel struct {
	Id_Role   uint   `gorm:"primaryKey;autoIncrement:true"`
	description *string
	CreatedAt    time.Time      
  	UpdatedAt    time.Time   
}