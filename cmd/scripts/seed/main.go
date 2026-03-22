package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"shopTemplate/app/db"
	"shopTemplate/app/models"
	"shopTemplate/plugins/auth"

	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	password := "password"
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal(err)
	}

	admin := auth.User{
		FirstName:       "Admin",
		LastName:        "User",
		Email:           "admin@botanica.com",
		PasswordHash:    string(hashedPassword),
		Role:            "admin",
		EmailVerifiedAt: sql.NullTime{Time: time.Now(), Valid: true},
	}

	// Check if admin exists
	var count int64
	db.Get().Model(&auth.User{}).Where("email = ?", admin.Email).Count(&count)
	if count == 0 {
		if err := db.Get().Create(&admin).Error; err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Admin created: %s / %s\n", admin.Email, password)
	} else {
		fmt.Printf("Admin already exists: %s\n", admin.Email)
	}

	user := auth.User{
		FirstName:       "John",
		LastName:        "Doe",
		Email:           "user@botanica.com",
		PasswordHash:    string(hashedPassword),
		Role:            "user",
		EmailVerifiedAt: sql.NullTime{Time: time.Now(), Valid: true},
	}

	// Check if user exists
	db.Get().Model(&auth.User{}).Where("email = ?", user.Email).Count(&count)
	if count == 0 {
		if err := db.Get().Create(&user).Error; err != nil {
			log.Fatal(err)
		}
		fmt.Printf("User created: %s / %s\n", user.Email, password)
	} else {
		fmt.Printf("User already exists: %s\n", user.Email)
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
