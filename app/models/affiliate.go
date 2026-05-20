package models

import "gorm.io/gorm"

type Affiliate struct {
	gorm.Model
	AffiliateID     string   `gorm:"uniqueIndex;size:20"`
	Name            string   `gorm:"size:100"`
	Email           string   `gorm:"size:255"`
	PasswordHash    string   `gorm:"size:255"`
	Rate            float64
	Active          bool     `gorm:"default:true"`
	APIToken        string   `gorm:"uniqueIndex;size:64"`
	ShopURL         string   `gorm:"size:255"`
	DashboardURL    string   `gorm:"size:255"`
	APIKey          string   `gorm:"size:128"`
	Balance         Currency `gorm:"type:numeric(12,2);default:100.00"`
	AuthorizedEmail string   `gorm:"size:255"`
}
