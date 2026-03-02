package controller

import (
	"log"
	"net/http"
	"time"

	"github.com/egutsenf96/warego/internal/database"
	"github.com/egutsenf96/warego/internal/models"
	"github.com/gin-gonic/gin"
)

func AddRole(c *gin.Context) {
	body := &models.Role{}
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
	data := models.Role{Description: body.Description, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	db.Create(&data)
	if db.Error != nil {
		pgl.Close()

		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to create role",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"data": data,
	})
}

func GetRoleById(c *gin.Context) {
	var role models.Role
	db, err := database.IntialDB()
	pgl, err := db.DB()

	if err != nil {
		log.Fatal(err)
		return
	}
	db.First(&role, c.Param("id"))
	pgl.Close()
	c.JSON(http.StatusOK, gin.H{
		"result": role,
	})
}

func GetAllRole(c *gin.Context) {
	var roles []models.Role
	db, err := database.IntialDB()
	pgl, err := db.DB()

	if err != nil {
		log.Fatal(err)
		return
	}
	db.Find(&roles)
	pgl.Close()
	c.JSON(http.StatusOK, gin.H{
		"result": roles,
	})
}

func UpdateRole(c *gin.Context) {
	body := &models.Role{}
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
	db.Model(&models.Role{}).Where("id_role = ?", c.Param("id")).Update("description", body.Description)
	if db.Error != nil {
		pgl.Close()

		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to update role",
		})
		return
	}
	pgl.Close()
	c.JSON(http.StatusOK, gin.H{
		"msg": "Update succesfully",
	})

}

func DeleteRole(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"msg": "Delete Role",
	})
}
