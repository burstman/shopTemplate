package handlers

import (
	"shopTemplate/app/db"
	"shopTemplate/app/models"
	"shopTemplate/app/views/landing"
	"shopTemplate/app/views/layouts"
	"strings"

	"github.com/anthdm/superkit/kit"
)

func HandleLandingIndex(kit *kit.Kit) error {
	// 1. Fetch all settings
	var settings []models.Setting
	err := db.Get().Find(&settings).Error
	if err != nil {
		return err
	}
	configMap := make(map[string]string)
	for _, s := range settings {
		configMap[s.Key] = s.Value
	}

	// 2. Fetch the specific products selected for the category section
	var categoryProducts []models.Product
	if idsStr := configMap["category_products"]; idsStr != "" {
		// Filter out empty strings to prevent SQL errors
		rawIDs := strings.Split(idsStr, ",")
		var ids []string
		for _, id := range rawIDs {
			if trimmed := strings.TrimSpace(id); trimmed != "" {
				ids = append(ids, trimmed)
			}
		}

		if len(ids) > 0 {
			err := db.Get().Where("id IN ?", ids).Find(&categoryProducts).Error
			if err != nil {
				return err
			}
		}
	}

	// 3. Fetch all products for the main shop display, ordered by newest first
	var allProducts []models.Product
	if err := db.Get().Order("created_at desc").Find(&allProducts).Error; err != nil {
		return err
	}

	user, _ := kit.Auth().(models.AuthUser)

	// Pass both featured and all products to the view
	return kit.Render(layouts.App(user, landing.Index(configMap, categoryProducts, allProducts)))
}
