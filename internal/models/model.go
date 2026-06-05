package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Base model
type Base struct {
	ID        uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

type Tenant struct {
	Base
	Name        string `gorm:"unique;not null" json:"name"`
	Domain      string `gorm:"unique;not null" json:"domain"`
	Description string `gorm:"type:text" json:"description,omitempty"`
	IsActive    bool   `gorm:"default:true" json:"is_active"`
	Users       []User `gorm:"foreignKey:TenantID" json:"users,omitempty"`
}

type Role struct {
	Base
	Name        string     `gorm:"not null;uniqueIndex:idx_name_tenant" json:"name"`
	Description string     `gorm:"type:text" json:"description,omitempty"`
	TenantID    *uuid.UUID `gorm:"type:uuid;index" json:"tenant_id,omitempty"`
	Tenant      *Tenant    `gorm:"foreignKey:TenantID" json:"tenant,omitempty"`
	IsGlobal    bool       `gorm:"default:false" json:"is_global,omitempty"`
	Users       []User     `gorm:"foreignKey:RoleID" json:"users,omitempty"`
}

type User struct {
	Base
	Name        string     `gorm:"not null" json:"name"`
	Email       string     `gorm:"unique;not null" json:"email"`
	Password    string     `gorm:"not null" json:"-"`
	RoleID      uuid.UUID  `gorm:"type:uuid;not null" json:"role_id"`
	TenantID    uuid.UUID  `gorm:"type:uuid;not null;index" json:"tenant_id"`
	ImageBase64 string     `gorm:"type:text" json:"image_base64,omitempty"`
	IsActive    bool       `gorm:"default:true" json:"is_active"`
	LastLogin   *time.Time `gorm:"type:timestamptz" json:"last_login,omitempty"`
	WarehouseID *uuid.UUID `gorm:"type:uuid;index" json:"warehouse_id,omitempty"`
	Role        *Role      `gorm:"foreignKey:RoleID" json:"role,omitempty"`
	Tenant      *Tenant    `gorm:"foreignKey:TenantID" json:"tenant,omitempty"`
	Warehouse   *Warehouse `gorm:"foreignKey:WarehouseID" json:"warehouse,omitempty"`
}

type Category struct {
	Base
	Name        string    `gorm:"not null" json:"name"`
	Description string    `gorm:"type:text" json:"description,omitempty"`
	TenantID    uuid.UUID `gorm:"type:uuid;not null;index" json:"tenant_id"`
	Tenant      Tenant    `gorm:"foreignKey:TenantID" json:"tenant,omitempty"`
	Products    []Product `gorm:"foreignKey:CategoryID" json:"products,omitempty"`
}

type Product struct {
	Base
	Name        string     `gorm:"not null" json:"name"`
	SKU         string     `gorm:"unique;not null" json:"sku"`
	Quantity    int        `gorm:"default:0" json:"quantity"`
	ImageBase64 string     `gorm:"type:text" json:"image_base64,omitempty"`
	CategoryID  uuid.UUID  `gorm:"type:uuid" json:"category_id"`
	TenantID    uuid.UUID  `gorm:"type:uuid;not null;index" json:"tenant_id"`
	Category    Category   `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
	Tenant      Tenant     `gorm:"foreignKey:TenantID" json:"tenant,omitempty"`
	SupplierID  *uuid.UUID `gorm:"type:uuid" json:"supplier_id,omitempty"`
}

type Supplier struct {
	Base
	Name     string    `gorm:"not null" json:"name"`
	Contact  string    `json:"contact,omitempty"`
	Email    string    `gorm:"unique" json:"email,omitempty"`
	Phone    string    `json:"phone,omitempty"`
	Address  string    `json:"address,omitempty"`
	TenantID uuid.UUID `gorm:"type:uuid;not null;index" json:"tenant_id"`
	Tenant   Tenant    `gorm:"foreignKey:TenantID" json:"tenant,omitempty"`
}
type Warehouse struct {
	Base
	Name     string    `gorm:"not null" json:"name"`
	Location string    `json:"location,omitempty"`
	TenantID uuid.UUID `gorm:"type:uuid;not null;index" json:"tenant_id"`
	Tenant   Tenant    `gorm:"foreignKey:TenantID" json:"tenant,omitempty"`
}

type Tracker struct {
	Base
	ProductID   uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_product_warehouse" json:"product_id"`
	WarehouseID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_product_warehouse" json:"warehouse_id"`
	Quantity    int       `gorm:"default:0" json:"quantity"`
	TenantID    uuid.UUID `gorm:"type:uuid;not null;index" json:"tenant_id"`
	Product     Product   `gorm:"foreignKey:ProductID" json:"product,omitempty"`
	Warehouse   Warehouse `gorm:"foreignKey:WarehouseID" json:"warehouse,omitempty"`
}

type StockTransactionType string

const (
	TypeInitial  StockTransactionType = "INITIAL"
	TypeReceive  StockTransactionType = "RECEIVE"
	TypePick     StockTransactionType = "PICK"
	TypeTransfer StockTransactionType = "TRANSFER"
	TypeAdjust   StockTransactionType = "ADJUST"
	TypeDraw     StockTransactionType = "DRAW"
)

type StockTransaction struct {
	Base
	ProductID      uuid.UUID            `gorm:"type:uuid;not null" json:"product_id"`
	WarehouseID    uuid.UUID            `gorm:"type:uuid;not null" json:"warehouse_id"`
	TenantID       uuid.UUID            `gorm:"type:uuid;not null;index" json:"tenant_id"`
	QuantityChange int                  `gorm:"not null" json:"quantity_change"`
	Type           StockTransactionType `gorm:"size:20;not null" json:"type"`
	ReferenceID    *uuid.UUID           `gorm:"type:uuid" json:"reference_id,omitempty"`
	UserID         *uuid.UUID           `gorm:"type:uuid" json:"user_id,omitempty"`
	SupplierID     *uuid.UUID           `gorm:"type:uuid" json:"supplier_id,omitempty"`
	Notes          string               `gorm:"type:text" json:"notes,omitempty"`
	Product        Product              `gorm:"foreignKey:ProductID" json:"product,omitempty"`
	Warehouse      Warehouse            `gorm:"foreignKey:WarehouseID" json:"warehouse,omitempty"`
	User           *User                `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Supplier       *Supplier            `gorm:"foreignKey:SupplierID" json:"supplier,omitempty"`
}

type AuditLogAction string

const (
	ActionLogin         AuditLogAction = "LOGIN"
	ActionLogout        AuditLogAction = "LOGOUT"
	ActionUserCreate    AuditLogAction = "USER_CREATE"
	ActionUserUpdate    AuditLogAction = "USER_UPDATE"
	ActionUserDelete    AuditLogAction = "USER_DELETE"
	ActionProductCreate AuditLogAction = "PRODUCT_CREATE"
	ActionProductUpdate AuditLogAction = "PRODUCT_UPDATE"
	ActionProductDelete AuditLogAction = "PRODUCT_DELETE"
	ActionStockAdjust   AuditLogAction = "STOCK_ADJUST"
	ActionTenantCreate  AuditLogAction = "TENANT_CREATE"
	ActionTenantUpdate  AuditLogAction = "TENANT_UPDATE"
)

type AuditLogChanges struct {
	Before map[string]interface{} `json:"before,omitempty"`
	After  map[string]interface{} `json:"after,omitempty"`
	Fields []string               `json:"fields"`
}

type AuditLog struct {
	Base
	UserID        uuid.UUID              `gorm:"type:uuid;index" json:"user_id"`
	UserEmail     string                 `gorm:"index" json:"user_email"`
	UserName      string                 `json:"user_name"`
	TargetType    string                 `gorm:"size:50;index" json:"target_type"`
	TargetID      *uuid.UUID             `gorm:"type:uuid;index" json:"target_id"`
	TargetName    string                 `json:"target_name,omitempty"`
	Action        AuditLogAction         `gorm:"size:50;not null;index" json:"action"`
	Description   string                 `gorm:"type:text" json:"description"`
	Changes       *AuditLogChanges       `gorm:"type:jsonb" json:"changes,omitempty"`
	TenantID      uuid.UUID              `gorm:"type:uuid;not null;index" json:"tenant_id"`
	IPAddress     string                 `gorm:"size:45" json:"ip_address,omitempty"`
	UserAgent     string                 `gorm:"type:text" json:"user_agent,omitempty"`
	RequestMethod string                 `gorm:"size:10" json:"request_method,omitempty"`
	RequestPath   string                 `gorm:"type:text" json:"request_path,omitempty"`
	Metadata      map[string]interface{} `gorm:"type:jsonb" json:"metadata,omitempty"`
}

func (AuditLog) TableName() string { return "audit_logs" }

func NewAuditLog(userID *uuid.UUID, userEmail, userName string, tenantID uuid.UUID,
	action AuditLogAction, targetType string, targetID *uuid.UUID,
	targetName, description string) *AuditLog {
	uid := uuid.Nil
	if userID != nil {
		uid = *userID
	}
	return &AuditLog{
		UserID:      uid,
		UserEmail:   userEmail,
		UserName:    userName,
		TenantID:    tenantID,
		Action:      action,
		TargetType:  targetType,
		TargetID:    targetID,
		TargetName:  targetName,
		Description: description,
		Metadata:    make(map[string]interface{}),
	}
}

func (a *AuditLog) WithChanges(before, after map[string]interface{}, fields []string) *AuditLog {
	if len(fields) == 0 {
		return a
	}
	a.Changes = &AuditLogChanges{Before: before, After: after, Fields: fields}
	return a
}

func (a *AuditLog) WithContext(ip, userAgent, method, path string) *AuditLog {
	a.IPAddress = ip
	a.UserAgent = userAgent
	a.RequestMethod = method
	a.RequestPath = path
	return a
}

func (a *AuditLog) Save(db *gorm.DB) error { return db.Create(a).Error }

type Draw struct {
	Base
	Name        string     `gorm:"not null" json:"name"`
	Description string     `gorm:"type:text" json:"description,omitempty"`
	ProductID   uuid.UUID  `gorm:"type:uuid;not null;index" json:"product_id"`
	Quantity    int        `gorm:"default:1;not null" json:"quantity"`
	WinnerID    *uuid.UUID `gorm:"type:uuid;index" json:"winner_id,omitempty"`
	Winner      *User      `gorm:"foreignKey:WinnerID" json:"winner,omitempty"`
	TenantID    uuid.UUID  `gorm:"type:uuid;not null;index" json:"tenant_id"`
	Tenant      Tenant     `gorm:"foreignKey:TenantID" json:"tenant,omitempty"`
	Status      string     `gorm:"size:20;default:'pending';index" json:"status"`
	RetrievedAt *time.Time `gorm:"type:timestamptz" json:"retrieved_at,omitempty"`
	RetrievedBy *uuid.UUID `gorm:"type:uuid" json:"retrieved_by,omitempty"`
	Retrieved   *User      `gorm:"foreignKey:RetrievedBy" json:"retrieved_user,omitempty"`
	Notes       string     `gorm:"type:text" json:"notes,omitempty"`
}

func (Draw) TableName() string { return "draws" }

type DrawStatus string

const (
	StatusPending   DrawStatus = "pending"
	StatusRetrieved DrawStatus = "retrieved"
	StatusCancelled DrawStatus = "cancelled"
)

func (d *Draw) IsRetrieved() bool    { return d.Status == string(StatusRetrieved) && d.RetrievedAt != nil }
func (d *Draw) CanBeCancelled() bool { return d.Status == string(StatusPending) }
func (d *Draw) MarkAsRetrieved(retrievedBy *uuid.UUID) {
	now := time.Now()
	d.Status = string(StatusRetrieved)
	d.RetrievedAt = &now
	d.RetrievedBy = retrievedBy
}
func (d *Draw) Cancel() error {
	if !d.CanBeCancelled() {
		return gorm.ErrInvalidValue
	}
	d.Status = string(StatusCancelled)
	return nil
}

type DrawQuery struct{ db *gorm.DB }

func NewDrawQuery(db *gorm.DB, tenantID uuid.UUID) *DrawQuery {
	return &DrawQuery{
		db: db.Where("tenant_id = ?", tenantID).Preload("Product").Preload("Winner").Preload("Retrieved"),
	}
}

func (q *DrawQuery) FilterByStatus(status DrawStatus) *DrawQuery {
	q.db = q.db.Where("status = ?", status)
	return q
}
func (q *DrawQuery) FilterByProduct(productID uuid.UUID) *DrawQuery {
	q.db = q.db.Where("product_id = ?", productID)
	return q
}
func (q *DrawQuery) FilterByWinner(winnerID uuid.UUID) *DrawQuery {
	q.db = q.db.Where("winner_id = ?", winnerID)
	return q
}
func (q *DrawQuery) FilterByDateRange(start, end time.Time) *DrawQuery {
	if !start.IsZero() {
		q.db = q.db.Where("created_at >= ?", start)
	}
	if !end.IsZero() {
		q.db = q.db.Where("created_at <= ?", end)
	}
	return q
}
func (q *DrawQuery) SearchByName(query string) *DrawQuery {
	if query != "" {
		q.db = q.db.Where("name ILIKE ?", "%"+query+"%")
	}
	return q
}
func (q *DrawQuery) OrderBy(field string, desc bool) *DrawQuery {
	if desc {
		q.db = q.db.Order(field + " DESC")
	} else {
		q.db = q.db.Order(field + " ASC")
	}
	return q
}
func (q *DrawQuery) Paginate(page, limit int) ([]Draw, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	var draws []Draw
	var total int64
	if err := q.db.Model(&Draw{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.db.Limit(limit).Offset((page - 1) * limit).Find(&draws).Error; err != nil {
		return nil, 0, err
	}
	return draws, total, nil
}
func (q *DrawQuery) Find() ([]Draw, error) { var d []Draw; return d, q.db.Find(&d).Error }
func (q *DrawQuery) First() (*Draw, error) {
	var d Draw
	if err := q.db.First(&d).Error; err != nil {
		return nil, err
	}
	return &d, nil
}
