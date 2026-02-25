package models

type DrawModel struct {
	Id_Draw uint `gorm:"primaryKey;autoIncrement:true"`
	Product_Id ProductModel `gorm:"foreignKey:Product_Id;references:Id_Product"`
	Stock float32
	Company_Id CompanyModel `gorm:"foreignKey:Company_Id;references:Id_Company"`
	User_Id UserModel `gorm:"foreignKey:User_Id;references:Id_User"`
	CreatedAt    time.Time      
  	
}