package migrations

import (
	"log"
	"net/http"

	"github.com/egutsenf96/warego/internal/database"
	"github.com/egutsenf96/warego/internal/models"
	"github.com/gin-gonic/gin"
)

func TrackerMigrationsUp(c *gin.Context) {
	db, err := database.IntialDB()
	if err != nil {
		log.Fatal(err)
	}
	err = db.Migrator().CreateTable(&models.Tracker{})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err,
		})
	}
	c.JSON(http.StatusAccepted, gin.H{
		"message": "Tracker migration execute succesfully",
	})
}

func TrackerMigrationsDown(c *gin.Context) {
	parameter := c.Query("delete")
	if parameter != "" {
		log.Fatal("Parameter not valid")
	}
	db, err := database.IntialDB()
	if err != nil {
		log.Fatal(err)
	}
	err = db.Migrator().DropTable(&models.Tracker{})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err,
		})
	}
	c.JSON(http.StatusAccepted, gin.H{
		"message": "Tracker migration execute succesfully",
	})
}
