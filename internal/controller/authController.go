package controller

import (
	"net/http"
	"os"
	"time"

	"github.com/egutsenf96/warego/internal/database"
	"github.com/egutsenf96/warego/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func SignUp(c *gin.Context) {
	var body struct {
		Name     string    `json:"name" binding:"required"`
		LastName string    `json:"last_name" binding:"required"`
		Email    string    `json:"email" binding:"required,email"`
		Password string    `json:"password" binding:"required,min=8"`
		TenantID uuid.UUID `json:"tenant_id" binding:"required"` // Multi-tenant requirement
		RoleID   uuid.UUID `json:"role_id"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Hash password
	hash, err := bcrypt.GenerateFromPassword([]byte(body.Password), 12)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	// Initialize DB (Ideally use a global instance instead of opening every time)
	db, _ := database.IntialDB()

	user := models.User{
		Base: models.Base{
			TenantID: body.TenantID,
		},
		FirstName: body.Name,
		LastName:  body.LastName,
		Email:     body.Email,
		Password:  string(hash),
		RoleID:    body.RoleID,
	}

	if result := db.Create(&user); result.Error != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "User already exists or invalid tenant"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "User created successfully", "user_id": user.ID})
}

func SingIn(c *gin.Context) {
	var body struct {
		Email    string `json:"email" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email and password required"})
		return
	}

	db, _ := database.IntialDB()
	var user models.User

	// 1. Find user by email
	if err := db.Where("email = ?", body.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	// 2. Compare password with hash
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(body.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	// 3. Generate JWT
	// We include TenantID in the token to persist context
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":       user.ID,
		"tenant_id": user.TenantID,
		"exp":       time.Now().Add(time.Hour * 24 * 7).Unix(),
	})

	tokenString, err := token.SignedString([]byte(os.Getenv("SECRETKEY")))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	// 4. Set Cookie & Response
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("Authorization", tokenString, 3600*24*7, "/", "", false, true)

	c.JSON(http.StatusOK, gin.H{
		"token":     tokenString,
		"tenant_id": user.TenantID, // Frontend will need this for X-Tenant-ID header
	})
}

func CheckAuth(c *gin.Context) {
	// Retrieving the user object set by your JwtValidate middleware
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": user,
	})
}

func GetAllUser(c *gin.Context) {
	var users []models.User
	db, err := database.IntialDB()
	pgl, err := db.DB()

	if err != nil {
		log.Fatal(err)
		return
	}
	db.Find(&users)
	pgl.Close()
	c.JSON(http.StatusOK, gin.H{
		"result": users,
	})
}
func GetUserById(c *gin.Context) {
	var user models.User
	db, err := database.IntialDB()
	pgl, err := db.DB()

	if err != nil {
		log.Fatal(err)
		return
	}
	db.First(&user, c.Param("id"))
	pgl.Close()
	c.JSON(http.StatusOK, gin.H{
		"result": user,
	})
}
func UpdateUser(c *gin.Context) {
	body := &models.User{}
	if c.Bind(&body) != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to read body",
		})
		return
	}
	db, err := database.IntialDB()
	pgl, err := db.DB()
	if err != nil {
		log.Fatal(err)
	}
	passwd, err := bcrypt.GenerateFromPassword([]byte(body.Password), 15)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"err": err,
		})
		return
	}
	db.Model(&models.User{}).Where("id_user = ?", c.Param("id")).Updates(models.User{Name: body.Name,
		LastName: body.LastName, Cargo: body.Cargo, Role_Id: body.Role_Id,
		Permisos: body.Permisos, Company_Id: body.Company_Id, Password: string(passwd), UpdatedAt: time.Now()})
	if db.Error != nil {
		pgl.Close()

		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to update user",
		})
		return
	}
	pgl.Close()
	c.JSON(http.StatusOK, gin.H{
		"msg": "Update succesfully",
	})

	var users []models.User

	db.Find(&users)
	pgl.Close()
	c.JSON(http.StatusOK, gin.H{
		"result": users,
	})
}
