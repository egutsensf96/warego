package main

import (
	"log"
	"net/http"

	"github.com/egutsenf96/warego/internal/controller"
	"github.com/egutsenf96/warego/internal/database/migrations"
	"github.com/egutsenf96/warego/internal/middleware"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
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
		auth.POST("/sing-in", controller.SingIn)
		auth.POST("/sign-up", controller.SignUp)
		auth.GET("/check", middleware.JwtValidate, controller.CheckAuth)

	}

	product := r.Group("/product")
	{
		product.GET("/")
		product.POST("/")
		product.PATCH("/")
		product.PUT("/")
		product.DELETE("/:id")
	}
	category := r.Group("/category")
	{
		category.GET("/")
		category.POST("/")
		category.PATCH("/")
		category.PUT("/")
		category.DELETE("/:id")
	}

	company := r.Group("/company")
	{
		company.GET("/")
		company.POST("/")
		company.PATCH("/")
		company.PUT("/")
		company.DELETE("/:id")
	}
	role := r.Group("/role")
	{
		role.GET("/", controller.GetAllRole)
		role.POST("/", controller.AddRole)
		role.PATCH("/:id", controller.UpdateRole)
		role.DELETE("/:id", controller.DeleteRole)
	}

	draw := r.Group("/draw")
	{
		draw.GET("/")
		draw.POST("/")
		draw.PATCH("/")
		draw.PUT("/")
		draw.DELETE("/:id")
	}

	r.GET("/sync", migrations.SchemaMigrations)

	r.Run() // listen and serve on
}
