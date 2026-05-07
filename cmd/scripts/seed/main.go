package main

import (
	"fmt"
	"log"

	"shopTemplate/app/db"
	"shopTemplate/app/models"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	if err := db.Connect(); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Define the settings with default values.
	// You can adjust the "Value" for each setting as needed.
	settings := []models.Setting{
		{Key: "carousel_count", Value: "5"},
		{Key: "category_count", Value: "4"},
		{Key: "best_seller_count", Value: "8"},
		{Key: "favorite_plants_count", Value: "4"},
	}

	for _, setting := range settings {
		// Use FirstOrCreate to prevent creating duplicate entries on subsequent runs.
		// It will check if a setting with the given key exists, and if not, it will create it.
		result := db.Get().Where(models.Setting{Key: setting.Key}).FirstOrCreate(&setting)
		if result.Error != nil {
			log.Printf("could not seed setting %q: %v\n", setting.Key, result.Error)
		}
	}

	// Seed Categories
	fmt.Println("Seeding categories...")

	// Locked Home category (cannot be deleted or modified)
	slug := "home"
	homeCat := models.Category{Name: "Home", Slug: &slug, IsLocked: true, Position: 0}
	db.Get().Where("slug = ?", "home").FirstOrCreate(&homeCat)

	// Top-level categories
	indoor := models.Category{Name: "Indoor Plants"}
	db.Get().FirstOrCreate(&indoor, models.Category{Name: "Indoor Plants"})

	outdoor := models.Category{Name: "Outdoor Plants"}
	db.Get().FirstOrCreate(&outdoor, models.Category{Name: "Outdoor Plants"})

	accessories := models.Category{Name: "Pots & Accessories"}
	db.Get().FirstOrCreate(&accessories, models.Category{Name: "Pots & Accessories"})

	// Sub-categories for Indoor Plants
	if indoor.ID > 0 {
		lowLight := models.Category{Name: "Low Light", ParentID: &indoor.ID}
		db.Get().FirstOrCreate(&lowLight, models.Category{Name: "Low Light", ParentID: &indoor.ID})

		petFriendly := models.Category{Name: "Pet Friendly", ParentID: &indoor.ID}
		db.Get().FirstOrCreate(&petFriendly, models.Category{Name: "Pet Friendly", ParentID: &indoor.ID})

		flowering := models.Category{Name: "Flowering", ParentID: &indoor.ID}
		db.Get().FirstOrCreate(&flowering, models.Category{Name: "Flowering", ParentID: &indoor.ID})
	}

	// Sub-categories for Pots & Accessories
	if accessories.ID > 0 {
		ceramic := models.Category{Name: "Ceramic Pots", ParentID: &accessories.ID}
		db.Get().FirstOrCreate(&ceramic, models.Category{Name: "Ceramic Pots", ParentID: &accessories.ID})

		stands := models.Category{Name: "Plant Stands", ParentID: &accessories.ID}
		db.Get().FirstOrCreate(&stands, models.Category{Name: "Plant Stands", ParentID: &accessories.ID})
	}

	log.Println("Database seeding finished.")

}
