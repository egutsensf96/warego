package controller

import (
	"net/http"
	"time"

	"github.com/egutsenf96/warego/internal/database"
	"github.com/egutsenf96/warego/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// GetAllEmployees retrieves all workers for the current tenant
func GetAllEmployees(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	db, _ := database.IntialDB()

	var employees []models.Employee

	// We preload "Contracts" so the frontend can see current salary/position info
	result := db.Preload("Contracts").Where("tenant_id = ?", tenantID).Find(&employees)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch employees"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"result": employees})
}

// CreateEmployee adds a new person to the HR database
func CreateEmployee(c *gin.Context) {
	tenantIDStr := c.GetString("tenantID")
	tenantID, _ := uuid.Parse(tenantIDStr)

	var body struct {
		FirstName string `json:"first_name" binding:"required"`
		LastName  string `json:"last_name" binding:"required"`
		Email     string `json:"email" binding:"required,email"`
		DNI       string `json:"dni" binding:"required"`
		Position  string `json:"position" binding:"required"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db, _ := database.IntialDB()

	employee := models.Employee{
		Base: models.Base{
			TenantID: tenantID,
		},
		FirstName: body.FirstName,
		LastName:  body.LastName,
		Email:     body.Email,
		DNI:       body.DNI,
		Position:  body.Position,
	}

	if err := db.Create(&employee).Error; err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Employee with this DNI or Email already exists"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": employee})
}

// GetEmployeeContracts returns all contracts across the company or for a specific employee
func GetEmployeeContracts(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	employeeID := c.Query("employee_id") // Optional filter via /contracts?employee_id=UUID

	db, _ := database.IntialDB()
	var contracts []models.Contract

	query := db.Where("tenant_id = ?", tenantID)

	if employeeID != "" {
		query = query.Where("employee_id = ?", employeeID)
	}

	if err := query.Find(&contracts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch contracts"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"result": contracts})
}

// SignNewContract creates a legal/financial agreement for an employee
func SignNewContract(c *gin.Context) {
	tenantIDStr := c.GetString("tenantID")
	tenantID, _ := uuid.Parse(tenantIDStr)

	var body struct {
		EmployeeID uuid.UUID  `json:"employee_id" binding:"required"`
		Type       string     `json:"type" binding:"required"` // e.g., "Permanent", "Internship"
		Salary     float64    `json:"salary" binding:"required,gt=0"`
		StartDate  time.Time  `json:"start_date" binding:"required"`
		EndDate    *time.Time `json:"end_date"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db, _ := database.IntialDB()

	// 1. Verify employee belongs to this tenant
	var emp models.Employee
	if err := db.Where("id = ? AND tenant_id = ?", body.EmployeeID, tenantID).First(&emp).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Employee not found in this company"})
		return
	}

	// 2. Create the contract
	contract := models.Contract{
		Base: models.Base{
			TenantID: tenantID,
		},
		EmployeeID: body.EmployeeID,
		Type:       body.Type,
		Salary:     body.Salary,
		StartDate:  body.StartDate,
		EndDate:    body.EndDate,
		IsActive:   true,
	}

	if err := db.Create(&contract).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to register contract"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": contract})
}
