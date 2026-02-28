package main

import (
	"log"
	"net/http"

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

	r.Use(func(c *gin.Context) {
		if c.Request.Host != "localhost:8080" {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid host header"})
			return
		}
		c.Header("X-Frame-Options", "DENY")
		c.Header("Content-Security-Policy", "default-src 'self'; connect-src *; font-src *; script-src-elem * 'unsafe-inline'; img-src * data:; style-src * 'unsafe-inline';")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		c.Header("Referrer-Policy", "strict-origin")
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("Permissions-Policy", "geolocation=(),midi=(),sync-xhr=(),microphone=(),camera=(),magnetometer=(),gyroscope=(),fullscreen=(self),payment=()")
		c.Next()
	})
	r.Use(gzip.Gzip(gzip.DefaultCompression))
	r.Use(cors.Default()) // All origins allowed by default

	auth := r.Group("/auth")
	{
		auth.POST("/login", login.GetLogin)
		auth.POST("/sign-up", signup.CreateUser)
	}

	r.GET("/sync", migrations.SchemaMigrations)

	r.Run() // listen and serve on
}
