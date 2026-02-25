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

	r.Group("/auth")
	{
		r.POST("/login", login.GetLogin)
		r.POST("/sign-up", signup.CreateUser)
	}

	r.Group("/sync")
	{
		r.GET("/schema", migrations.UserMigrationsUP)
		r.POST("/schema", migrations.UserMigrationsDown)
	}

	r.Run() // listen and serve on
}
