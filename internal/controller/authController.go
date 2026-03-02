package controller

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/egutsenf96/warego/internal/database"
	"github.com/egutsenf96/warego/internal/models"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/gin-gonic/gin"
)

func SignUp(c *gin.Context) {
	body := &models.User{}
	db, err := database.IntialDB()
	pgl, err := db.DB()
	if err != nil {
		log.Fatal(err)
	}

	if c.Bind(&body) != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to read body",
		})
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(body.Password), 15)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to hash password",
		})
		return
	}
	user := models.User{Name: body.Name, LastName: body.LastName, Cargo: body.Cargo,
		Permisos: body.Permisos, Email: body.Email, Password: string(hash),
		Company_Id: body.Company_Id, Role_Id: body.Role_Id, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	db.Create(&user)
	pgl.Close()
	if db.Error != nil {

		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to create user",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"data": user,
	})
}

func SingIn(c *gin.Context) {
	var body struct {
		email    string
		password string
	}
	if c.Bind(&body) != nil {
		c.JSON(400, gin.H{
			"message": "Invalid body",
		})
		return
	}
	db, err := database.IntialDB()
	pgl, err := db.DB()
	if err != nil {
		log.Fatal(err)
	}
	db.First(&models.User{}, "email = ?", body.email)
	err = bcrypt.CompareHashAndPassword([]byte(models.User{}.Password), []byte(body.password))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid email or password",
		})
		return
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, jwt.MapClaims{
		"token": models.User{}.Id_User,
		"exp":   time.Now().Add(time.Hour * 24 * 7).Unix(),
	})

	tokenString, err := token.SignedString(os.Getenv("SECRETKEY"))
	pgl.Close()
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("Authorization", tokenString, 3600*24*7, "/", "", false, true)
	c.JSON(http.StatusAccepted, gin.H{
		"mesg": "ok",
	})
}

func CheckAuth(c *gin.Context) {
	user, _ := c.Get("user")
	c.JSON(http.StatusOK, gin.H{
		"msg": user,
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
