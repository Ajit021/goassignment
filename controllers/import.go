package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"goassigment/config"
	"goassigment/models"
	"log"
	"net/http"
	"regexp"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
)

func UploadExcel(c *gin.Context) {
	// Get the uploaded file
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to upload file: " + err.Error()})
		return
	}

	// Save the uploaded file temporarily
	filePath := "./uploads/" + file.Filename
	if err := c.SaveUploadedFile(file, filePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file: " + err.Error()})
		return
	}

	// Parse and process the Excel file asynchronously
	go processExcelFile(filePath)

	// Respond to the client
	c.JSON(http.StatusOK, gin.H{"message": "File is being processed"})
}

func processExcelFile(filePath string) {
	// Open the Excel file
	xlsx, err := excelize.OpenFile(filePath)
	if err != nil {
		log.Printf("Error opening Excel file: %v", err)
		return
	}

	// Get sheet list and read rows from the first sheet
	sheets := xlsx.GetSheetList()
	if len(sheets) == 0 {
		log.Printf("Error: No sheets found in the Excel file")
		return
	}
	sheetName := sheets[0]
	rows, err := xlsx.GetRows(sheetName)
	if err != nil {
		log.Printf("Error reading rows from Excel: %v", err)
		return
	}

	// Validate data headers and row data
	err = validateExcelData(rows)
	if err != nil {
		log.Printf("Error validating Excel data: %v", err)
		return
	}

	// Asynchronous processing using channels for error handling
	errChannel := make(chan error)

	// Insert data into MySQL asynchronously
	go func() {
		for _, row := range rows[1:] { // Skip header row
			err := insertDataIntoDB(row)
			if err != nil {
				errChannel <- err
				return
			}
		}
		errChannel <- nil
	}()

	// Cache data into Redis asynchronously
	go func() {
		err := cacheDataInRedis(rows[1:])
		if err != nil {
			errChannel <- err
			return
		}
		errChannel <- nil
	}()

	// Wait for both operations to finish
	for i := 0; i < 2; i++ {
		err := <-errChannel
		if err != nil {
			log.Printf("Error during processing: %v", err)
			return
		}
	}

	log.Println("File processed successfully")
}

// Validate headers and data types from the Excel rows
func validateExcelData(rows [][]string) error {
	// Define expected column headers
	expectedHeaders := []string{"FirstName", "LastName", "Company", "Address", "City", "Country", "Postal", "Phone", "Email", "Web"}
	headers := rows[0] // The first row is assumed to be the header

	// Validate header columns
	for i, header := range headers {
		if header != expectedHeaders[i] {
			return fmt.Errorf("invalid header at column %d: expected '%s', got '%s'", i+1, expectedHeaders[i], header)
		}
	}

	emailPattern := `^[a-z0-9._%+-]+@[a-z0-9.-]+\.[a-z]{2,}$`
	emailRegex, err := regexp.Compile(emailPattern)
	if err != nil {
		return fmt.Errorf("failed to compile email regex: %v", err)
	}

	// Validate data types (example: check if phone number is valid)
	for _, row := range rows[1:] { // Skip header row
		if len(row) != len(expectedHeaders) {
			return fmt.Errorf("row has incorrect number of columns: %v", row)
		}

		// Validate email format (check using the precompiled regex)
		if !emailRegex.MatchString(row[8]) {
			return fmt.Errorf("invalid email format in row: %v", row)
		}
	}

	return nil
}

// Insert employee data into MySQL database
func insertDataIntoDB(row []string) error {
	employee := models.Employee{
		FirstName: row[0],
		LastName:  row[1],
		Company:   row[2],
		Address:   row[3],
		City:      row[4],
		Country:   row[5],
		Postal:    row[6],
		Phone:     row[7],
		Email:     row[8],
		Web:       row[9],
	}

	// Insert data into MySQL
	if err := config.DB.Create(&employee).Error; err != nil {
		return fmt.Errorf("error inserting data into MySQL: %v", err)
	}

	return nil
}

// Cache employee data into Redis
func cacheDataInRedis(rows [][]string) error {
	ctx := context.Background()
	var employees []models.Employee

	// Prepare employee data for caching
	for _, row := range rows {
		employee := models.Employee{
			FirstName: row[0],
			LastName:  row[1],
			Company:   row[2],
			Address:   row[3],
			City:      row[4],
			Country:   row[5],
			Postal:    row[6],
			Phone:     row[7],
			Email:     row[8],
			Web:       row[9],
		}
		employees = append(employees, employee)
	}

	// Marshal the employee data to JSON
	data, _ := json.Marshal(employees)

	// Store in Redis with an expiration of 5 minutes
	err := config.RedisClient.Set(ctx, "employee_data:", data, 5*time.Minute).Err()
	if err != nil {
		fmt.Println("Error caching data:", err)
	}

	return nil
}
