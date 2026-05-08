package handlers

import (
	"crypto/rand"
	"database/sql"
	"fmt"
	"math/big"
	"net/http"
	"shopTemplate/app/db"
	"shopTemplate/app/models"
	"shopTemplate/app/views/admin"
	"time"

	"github.com/anthdm/superkit/kit"
	"golang.org/x/crypto/bcrypt"
)

var randomChars = []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789")

func randomString(length int) (string, error) {
	b := make([]rune, length)
	for i := range b {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(randomChars))))
		if err != nil {
			return "", err
		}
		b[i] = randomChars[n.Int64()]
	}
	return string(b), nil
}

var generateSetupPassword = randomString

func HandleSetupIndex(kit *kit.Kit) error {
	var count int64
	db.Get().Model(&models.User{}).Where("role = ?", "admin").Count(&count)
	if count > 0 {
		return kit.Redirect(http.StatusSeeOther, "/login")
	}
	return kit.Render(admin.SetupPage("", "", "", ""))
}

func HandleSetupCreate(kit *kit.Kit) error {
	var count int64
	db.Get().Model(&models.User{}).Where("role = ?", "admin").Count(&count)
	if count > 0 {
		return kit.Redirect(http.StatusSeeOther, "/login")
	}

	name := kit.Request.FormValue("name")
	email := kit.Request.FormValue("email")

	if name == "" || email == "" {
		return kit.Render(admin.SetupPage("", "", "", "All fields are required."))
	}

	scheme := "https"
	if kit.Request.TLS == nil {
		scheme = "http"
	}
	domain := fmt.Sprintf("%s://%s", scheme, kit.Request.Host)

	password, err := generateSetupPassword(12)
	if err != nil {
		return err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user := models.User{
		Email:           email,
		FirstName:       name,
		PasswordHash:    string(hash),
		Role:            "admin",
		EmailVerifiedAt: sql.NullTime{Time: time.Now(), Valid: true},
	}

	if err := db.Get().Create(&user).Error; err != nil {
		return kit.Render(admin.SetupPage("", "", "", "Failed to create admin: "+err.Error()))
	}

	// Auto-create affiliate only if one doesn't exist
	var affCount int64
	db.Get().Model(&models.Affiliate{}).Count(&affCount)
	if affCount == 0 {
		affiliate := models.Affiliate{
			AffiliateID:  "AFF-001",
			Name:         name,
			Email:        email,
			PasswordHash: string(hash),
			Rate:         0,
			Active:       true,
			Domain:       domain,
		}
		if err := db.Get().Create(&affiliate).Error; err != nil {
			return kit.Render(admin.SetupPage("", "", "", "Failed to create affiliate: "+err.Error()))
		}
	}

	return kit.Render(admin.SetupPage(email, password, domain, ""))
}
