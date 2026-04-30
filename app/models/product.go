package models

import "gorm.io/gorm"

type Product struct {
	gorm.Model
	Name           string
	Price          Currency `gorm:"type:numeric(12,2)"`
	PromotionPrice Currency `gorm:"type:numeric(12,2)"`
	Description    string
	Stock          int
	Image          string
	Categories     []Category `gorm:"many2many:product_categories;"`
	Category       string     `gorm:"-"` // Deprecated: Kept for backward compatibility with views
}
