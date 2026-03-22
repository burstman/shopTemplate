package models

import "gorm.io/gorm"

type Product struct {
	gorm.Model
	Name       string
	Price      float64
	Image      string
	Categories []Category `gorm:"many2many:product_categories;"`
	Category   string     `gorm:"-"` // Deprecated: Kept for backward compatibility with views
}
