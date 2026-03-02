package migrations

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func SchemaMigrations(c *gin.Context) {
	schema := c.Query("schema")
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
		return
	}

}
