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

func GetLogin(c *gin.Context) {
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
