package models

import "gorm.io/gorm"

type Affiliate struct {
	gorm.Model
	AffiliateID  string `gorm:"uniqueIndex;size:20"`
	Name         string `gorm:"size:100"`
	Email        string `gorm:"size:255"`
	PasswordHash string `gorm:"size:255"`
	Rate         float64
	Active       bool   `gorm:"default:true"`
	APIToken     string `gorm:"uniqueIndex;size:64"`
	Domain       string `gorm:"size:255"`
}
