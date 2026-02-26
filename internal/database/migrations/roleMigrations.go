package migrations

import (
	"log"
	"net/http"

	"github.com/egutsenf96/warego/internal/database"
	"github.com/egutsenf96/warego/internal/models"
	"github.com/gin-gonic/gin"
)

func RoleMigrationsUp(c *gin.Context) {
	db, err := database.IntialDB()
	if err != nil {
		log.Fatal(err)
	}
	err = db.Migrator().CreateTable(&models.Role{})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err,
		})
	}
	c.JSON(http.StatusAccepted, gin.H{
		"message": "Role migration execute succesfully",
	})
}

func RoleMigrationsDown(c *gin.Context) {
	parameter := c.Query("delete")
	if parameter != "" {
		log.Fatal("Parameter not valid")
	}
	db, err := database.IntialDB()
	if err != nil {
		log.Fatal(err)
	}
	err = db.Migrator().DropTable(&models.Role{})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err,
		})
	}
	c.JSON(http.StatusAccepted, gin.H{
		"message": "Role migration execute succesfully",
	})
}
