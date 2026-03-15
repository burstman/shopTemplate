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

	log.Println("Database seeding finished.")

}
