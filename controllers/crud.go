package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"goassigment/config"
	"goassigment/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func GetImportedData(c *gin.Context) {
	ctx := context.Background()

	// Attempt to fetch data from Redis
	data, err := config.RedisClient.Get(ctx, "employee_data").Result()
	fmt.Println("Redis Error->", err)

	if err != nil {
		// Cache miss, data not found in Redis
		fmt.Println("Cache miss: Fetching from MySQL")

		// Fetch data from MySQL
		var employees []models.Employee
		if err := config.DB.Find(&employees).Error; err != nil {
			// Return an error response if fetching from DB fails
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch data from DB"})
			return
		}

		// Marshal the employees data to JSON (this returns []byte)
		dataBytes, err := json.Marshal(employees)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error marshaling data", "details": err.Error()})
			return
		}

		// Convert []byte to string (this will be stored in Redis)
		dataString := string(dataBytes)

		// Cache the data in Redis for 5 minutes
		err = config.RedisClient.Set(ctx, "employee_data", dataString, 5*time.Minute).Err()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error caching data in Redis", "details": err.Error()})
			return
		}

		// Return the fetched data as a JSON response
		c.JSON(http.StatusOK, gin.H{"data": json.RawMessage(dataString)})
		return
	}

	// // If no error or an error other than redis.Nil occurs
	// if err != nil {
	// 	c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching data from Redis", "details": err.Error()})
	// 	return
	// }

	// If data was found in Redis, return it
	fmt.Println("Cache hit: Returning data from Redis")
	c.JSON(http.StatusOK, gin.H{"data": json.RawMessage(data)})
}

func EditRecord(c *gin.Context) {
	id := c.Param("id")
	var input models.Employee
	if err := c.BindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	var employee models.Employee
	if err := config.DB.First(&employee, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Record not found"})
		return
	}

	// Update MySQL
	config.DB.Model(&employee).Updates(input)

	// Update Redis
	ctx := context.Background()
	var employees []models.Employee
	config.DB.Find(&employees)
	data, _ := json.Marshal(employees)
	config.RedisClient.Set(ctx, "employee_data", data, 5*time.Minute)

	c.JSON(http.StatusOK, gin.H{"message": "Record updated"})
}
