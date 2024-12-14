package main

import (
	"goassigment/config"
	"goassigment/controllers"
	"goassigment/models"

	"github.com/gin-gonic/gin"
)

func main() {
	// Initialize DB and Redis
	config.InitDB()
	config.InitRedis()

	// Auto-migrate schema
	config.DB.AutoMigrate(&models.Employee{})

	// Create router
	r := gin.Default()

	// Routes
	r.POST("/upload", controllers.UploadExcel)
	r.GET("/employees", controllers.GetImportedData)
	r.PUT("/employee/:id", controllers.EditRecord)

	// Start server
	r.Run(":8080")
}
