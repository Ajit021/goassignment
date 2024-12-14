package models

import "gorm.io/gorm"

type Employee struct {
	gorm.Model
	FirstName string
	LastName  string
	Company   string
	Address   string
	City      string
	Country   string
	Postal    string
	Phone     string
	Email     string
	Web       string
}
