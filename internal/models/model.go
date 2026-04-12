package models

import (
	"time"
)

// Tenant defines the top-level Organization (Multi-tenant)
type Tenant struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	TaxID     string    `json:"tax_id"`
	CreatedAt time.Time `json:"created_at"`
}

// Category allows for product grouping
type Category struct {
	ID            int        `json:"id"`
	TenantID      int        `json:"tenant_id"`
	Name          string     `json:"name"`
	ParentID      *int       `json:"parent_id"` // Pointer for nullable parent
	SubCategories []Category `json:"sub_categories,omitempty"`
}

type Base struct {
	ID        int       `gorm:"primaryKey" json:"id"`
	TenantID  int       `gorm:"index" json:"tenant_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type User struct {
	Base                // This embeds ID and TenantID
	Username     string `gorm:"unique" json:"username"`
	Email        string `gorm:"unique" json:"email"`
	PasswordHash string `json:"-"` // This is the field the controller is looking for
	RoleID       int    `json:"role_id"`
	Role         Role   `json:"role" gorm:"foreignKey:RoleID"`
	ImageBase64  string `json:"image_base64"`
	IsActive     bool   `json:"is_active" gorm:"default:true"`
}

type Role struct {
	ID       int    `gorm:"primaryKey" json:"id"`
	TenantID int    `json:"tenant_id"`
	Name     string `json:"name"`
}

// ProductTemplate (The "Base" Product)
type ProductTemplate struct {
	ID          int              `json:"id"`
	TenantID    int              `json:"tenant_id"`
	CategoryID  int              `json:"category_id"`
	Category    Category         `json:"category"` // Belongs To
	Name        string           `json:"name"`
	Type        string           `json:"type"`
	ImageBase64 string           `json:"image_base64"`
	Variants    []ProductVariant `json:"variants"` // Has Many
}

// ProductVariant (The specific item, e.g., T-Shirt Blue Small)
type ProductVariant struct {
	ID              int            `json:"id"`
	TemplateID      int            `json:"template_id"`
	TenantID        int            `json:"tenant_id"`
	SKU             string         `json:"sku"`
	AttributeValues map[string]any `json:"attribute_values"` // JSONB mapping
	CostPrice       float64        `json:"cost_price"`
	ListPrice       float64        `json:"list_price"`
	Active          bool           `json:"active"`
}

// Multi-Warehouse / Location Relationship
type Warehouse struct {
	ID        int        `json:"id"`
	TenantID  int        `json:"tenant_id"`
	Name      string     `json:"name"`
	Code      string     `json:"code"`
	Locations []Location `json:"locations"` // Has Many
}

type Location struct {
	ID           int    `json:"id"`
	WarehouseID  int    `json:"warehouse_id"`
	TenantID     int    `json:"tenant_id"`
	Name         string `json:"name"`
	LocationType string `json:"location_type"`
}

// StockMove tracks movement between locations
type StockMove struct {
	ID             int            `json:"id"`
	TenantID       int            `json:"tenant_id"`
	VariantID      int            `json:"variant_id"`
	Variant        ProductVariant `json:"variant"` // Belongs To
	SrcLocationID  int            `json:"src_location_id"`
	DestLocationID int            `json:"dest_location_id"`
	UserID         int            `json:"user_id"`
	User           User           `json:"user"` // Belongs To
	Qty            float64        `json:"qty"`
	Reference      string         `json:"reference"`
	CreatedAt      time.Time      `json:"created_at"`
}
type Tracker struct {
	Base             // Embeds ID, TenantID, CreatedAt
	UserID    int    `json:"user_id" gorm:"index"`          // The person who performed the action
	User      User   `json:"user" gorm:"foreignKey:UserID"` // Belongs To relationship
	Event     string `json:"event"`                         // e.g., "PRODUCT_CREATED", "STOCK_DRAWAL"
	Resource  string `json:"resource"`                      // e.g., "product_variants", "stock_moves"
	TargetID  int    `json:"target_id"`                     // The ID of the record that was changed
	OldValue  string `json:"old_value" gorm:"type:text"`    // For tracking changes
	NewValue  string `json:"new_value" gorm:"type:text"`    // For tracking changes
	IPAddress string `json:"ip_address"`                    // Useful for security audits
}
