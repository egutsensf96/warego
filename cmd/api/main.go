package main

import (
	"log"

	"github.com/egutsenf96/warego/internal/controller/login"
	"github.com/egutsenf96/warego/internal/controller/signup"
	"github.com/egutsenf96/warego/internal/database"
	"github.com/egutsenf96/warego/internal/migrations"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	database.IntialDB() //Open DB conection
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	r := gin.Default()
	r.Use(gzip.Gzip(gzip.DefaultCompression))
	r.Use(cors.Default()) // All origins allowed by default

	auth := r.Group("/auth")
	{
		auth.POST("/login", login.GetLogin)
		auth.POST("/sign-up", signup.CreateUser)
	}

	sync := r.Group("/sync")
	{
		sync.GET("/schema/role", migrations.RoleMigrationsUp)
		sync.POST("/schema/role", migrations.RoleMigrationsDown)
		sync.GET("/schema/track", migrations.TrackerMigrationsUp)
		sync.POST("/schema/track", migrations.TrackerMigrationsDown)
		sync.GET("/schema/company", migrations.CompanyMigrationsUp)
		sync.POST("/schema/company", migrations.CompanyMigrationsDown)
		sync.GET("/schema/category", migrations.CategoryMigrationUp)
		sync.POST("/schema/category", migrations.CategoryMigrationDown)
		sync.GET("/schema/draw", migrations.DrawMigrationsUp)
		sync.POST("/schema/draw", migrations.DrawMigrationsDown)
		sync.GET("/schema/product", migrations.ProductMigrationsUp)
		sync.POST("/schema/product", migrations.ProductMigrationsDown)
		sync.GET("/schema/user", migrations.UserMigrationsUP)
		sync.POST("/schema/user", migrations.UserMigrationsDown)

	}

	r.Run() // listen and serve on
}
