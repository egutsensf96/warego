package models

import "gorm.io/gorm"

type UserModel struct {
	gorm.Model
	Nombre   *string
	Apellido *string
	email    string
	password string
}
