package models

import "gorm.io/gorm"

type Order struct {
	gorm.Model
	FirstName string
	LastName  string
	Email     string
	Address   string
	City      string
	Phone     string
	Total     float64
	Status    string // pending, completed, cancelled
	Items     []OrderItem
}

type OrderItem struct {
	gorm.Model
	OrderID      uint
	ProductID    uint
	Product      Product
	ProductName  string
	ProductImage string
	Quantity     int
	Price        float64
}
