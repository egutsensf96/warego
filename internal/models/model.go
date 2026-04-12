package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Base model to include common fields
type Base struct {
	ID        uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

type Tenant struct {
	Base
	Name   string `gorm:"unique;not null" json:"name"`
	Domain string `gorm:"unique" json:"domain"`
}

type Role struct {
	Base
	Name        string `gorm:"not null" json:"name"`
	Description string `json:"description"`
}

type User struct {
	Base
	Email       string    `gorm:"unique;not null" json:"email"`
	Password    string    `gorm:"not null" json:"-"`
	ImageBase64 string    `gorm:"type:text" json:"image_base64"`
	TenantID    uuid.UUID `gorm:"type:uuid" json:"tenant_id"`
	RoleID      uuid.UUID `gorm:"type:uuid" json:"role_id"`
	Tenant      Tenant    `gorm:"foreignKey:TenantID" json:"tenant,omitempty"`
	Role        Role      `gorm:"foreignKey:RoleID" json:"role,omitempty"`
}

type Category struct {
	Base
	Name     string    `gorm:"not null" json:"name"`
	TenantID uuid.UUID `gorm:"type:uuid" json:"tenant_id"`
	Tenant   Tenant    `gorm:"foreignKey:TenantID" json:"tenant,omitempty"`
}

type Product struct {
	Base
	Name        string    `gorm:"not null" json:"name"`
	SKU         string    `gorm:"unique;not null" json:"sku"`
	ImageBase64 string    `gorm:"type:text" json:"image_base64"`
	CategoryID  uuid.UUID `gorm:"type:uuid" json:"category_id"`
	TenantID    uuid.UUID `gorm:"type:uuid" json:"tenant_id"`
	Category    Category  `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
	Tenant      Tenant    `gorm:"foreignKey:TenantID" json:"tenant,omitempty"`
}

type Warehouse struct {
	Base
	Name     string    `gorm:"not null" json:"name"`
	Location string    `json:"location"`
	TenantID uuid.UUID `gorm:"type:uuid" json:"tenant_id"`
	Tenant   Tenant    `gorm:"foreignKey:TenantID" json:"tenant,omitempty"`
}

type Tracker struct {
	Base
	ProductID   uuid.UUID `gorm:"type:uuid" json:"product_id"`
	WarehouseID uuid.UUID `gorm:"type:uuid" json:"warehouse_id"`
	Quantity    int       `json:"quantity"`
	TenantID    uuid.UUID `gorm:"type:uuid" json:"tenant_id"`
	Product     Product   `gorm:"foreignKey:ProductID" json:"product,omitempty"`
	Warehouse   Warehouse `gorm:"foreignKey:WarehouseID" json:"warehouse,omitempty"`
}

type Draw struct {
	Base
	Name          string     `json:"name"`
	ProductID     uuid.UUID  `gorm:"type:uuid" json:"product_id"`
	WinnerID      *uuid.UUID `gorm:"type:uuid" json:"winner_id"`
	TenantID      uuid.UUID  `gorm:"type:uuid" json:"tenant_id"`
	RetrievedByID *uuid.UUID `gorm:"type:uuid" json:"retrieved_by_id"`
	RetrievedAt   *time.Time `json:"retrieved_at"`
	Status        string     `gorm:"default:'pending'" json:"status"`
	Product       Product    `gorm:"foreignKey:ProductID" json:"product,omitempty"`
	Winner        User       `gorm:"foreignKey:WinnerID" json:"winner,omitempty"`
	RetrievedBy   User       `gorm:"foreignKey:RetrievedByID" json:"retrieved_by,omitempty"`
	Tenant        Tenant     `gorm:"foreignKey:TenantID" json:"tenant,omitempty"`
}
