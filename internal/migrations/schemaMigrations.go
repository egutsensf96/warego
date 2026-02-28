package migrations

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func SchemaMigrations(c *gin.Context) {
	schema := c.Query("schema")
	delete := c.Query("delete")
	if schema == "" {
		if delete == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "Parameter not allowed",
			})

		}
	}
	if schema != "true" || delete != "true" {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Value not found",
		})
	}
	if schema == "true" {
		RoleMigrationsUp()
		CategoryMigrationUp()
		CompanyMigrationsUp()
		UserMigrationsUP()
		ProductMigrationsUp()
		DrawMigrationsUp()
		TrackerMigrationsUp()
		time.Sleep(time.Millisecond * 4)
		c.JSON(http.StatusAccepted, gin.H{
			"message": "Execute successfully all migrations",
		})
	}
	if delete == "true" {
		TrackerMigrationsDown()
		DrawMigrationsDown()
		ProductMigrationsDown()
		UserMigrationsDown()
		CompanyMigrationsDown()
		CategoryMigrationDown()
		RoleMigrationsDown()
		c.JSON(http.StatusAccepted, gin.H{
			"message": "Execute successfully all migrations",
		})
	}
}
