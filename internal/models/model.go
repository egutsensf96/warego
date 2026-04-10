package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Base struct {
	ID        uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	TenantID  uuid.UUID      `gorm:"type:uuid;index;not null"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

type Company struct {
	Base
	Name        string `json:"name"`
	TaxID       string `json:"tax_id"` // e.g., RIF or NIT
	Description string `json:"description"`
}

type Role struct {
	Base
	Description string `json:"description"` // e.g., "Admin", "Inventory Clerk"
}

type Draw struct {
	Base
	Product_Id uuid.UUID `gorm:"type:uuid"`
	Product    Product   `gorm:"foreignKey:Product_Id"`
	Stock      float32
	User_Id    uuid.UUID `gorm:"type:uuid"`
	User       User      `gorm:"foreignKey:User_Id"`
}

type User struct {
	Base
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Email     string    `gorm:"uniqueIndex:idx_tenant_user_email" json:"email"`
	Password  string    `json:"-"`
	RoleID    uuid.UUID `gorm:"type:uuid"`
	Role      Role      `gorm:"foreignKey:RoleID"`
}

type Category struct {
	Base
	Description string `json:"description"`
}

type Warehouse struct {
	Base
	Name     string `json:"name"`
	Location string `json:"location"`
}

type Product struct {
	Base
	SKU         string    `gorm:"uniqueIndex:idx_tenant_sku" json:"sku"`
	Description string    `json:"description"`
	Cost        float64   `json:"cost"`
	CategoryID  uuid.UUID `gorm:"type:uuid"`
	Category    Category  `gorm:"foreignKey:CategoryID"`
	CreatedBy   uuid.UUID `gorm:"type:uuid"`
	UserID      uuid.UUID // This is the field name Go expects
	User        User      `gorm:"foreignKey:CreatedBy"`
}

type Stock struct {
	Base
	ProductID   uuid.UUID `gorm:"type:uuid"`
	Product     Product   `gorm:"foreignKey:ProductID"`
	WarehouseID uuid.UUID `gorm:"type:uuid"`
	Warehouse   Warehouse `gorm:"foreignKey:WarehouseID"`
	Quantity    float64   `json:"quantity"`
}

type Employee struct {
	Base
	FirstName string     `json:"first_name"`
	LastName  string     `json:"last_name"`
	Email     string     `json:"email"`
	DNI       string     `gorm:"uniqueIndex:idx_tenant_dni" json:"dni"`
	Position  string     `json:"position"`
	Contracts []Contract `gorm:"foreignKey:EmployeeID"`
}

type Contract struct {
	Base
	EmployeeID uuid.UUID  `gorm:"type:uuid"`
	Type       string     `json:"type"` // e.g., "Full-Time", "Freelance"
	Salary     float64    `json:"salary"`
	StartDate  time.Time  `json:"start_date"`
	EndDate    *time.Time `json:"end_date,omitempty"`
	IsActive   bool       `gorm:"default:true" json:"is_active"`
}

type Tracker struct {
	Base
	UserID    uuid.UUID `gorm:"type:uuid"`
	User      User      `gorm:"foreignKey:UserID"`
	Event     string    `json:"event"`     // e.g., "PRODUCT_OUT", "CONTRACT_VOID"
	TargetID  uuid.UUID `gorm:"type:uuid"` // ID of the object changed
	OldValues string    `gorm:"type:text"` // JSON string of old data
	NewValues string    `gorm:"type:text"` // JSON string of new data
}
