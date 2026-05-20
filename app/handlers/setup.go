package handlers

import (
	"crypto/rand"
	"database/sql"
	"fmt"
	"math/big"
	"net/http"
	"shopTemplate/app/config"
	"shopTemplate/app/db"
	"shopTemplate/app/models"
	"shopTemplate/app/views/admin"
	"strconv"
	"strings"
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

func generateAffiliateID() (string, error) {
	var maxID *string
	err := db.Get().Model(&models.Affiliate{}).Select("MAX(affiliate_id)").Scan(&maxID).Error
	if err != nil {
		return "", err
	}
	if maxID == nil || *maxID == "" {
		return "AFF-001", nil
	}
	parts := strings.SplitN(*maxID, "-", 2)
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid affiliate id format: %s", *maxID)
	}
	num, err := strconv.Atoi(parts[1])
	if err != nil {
		return "", fmt.Errorf("invalid affiliate id number: %s", parts[1])
	}
	return fmt.Sprintf("AFF-%03d", num+1), nil
}

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

	// Check if the email is authorized
	var authorizedEmails []string
	db.Get().Model(&models.Affiliate{}).Where("authorized_email <> ''").Pluck("authorized_email", &authorizedEmails)
	if len(authorizedEmails) > 0 {
		authorized := false
		for _, ae := range authorizedEmails {
			if ae == email {
				authorized = true
				break
			}
		}
		if !authorized {
			return kit.Render(admin.SetupPage("", "", "", "Your email is not authorized for setup. Contact the shop owner."))
		}
	}

	scheme := "https"
	if kit.Request.TLS == nil {
		scheme = "http"
	}
	shopURL := fmt.Sprintf("%s://%s", scheme, kit.Request.Host)

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
		affiliateID, err := generateAffiliateID()
		if err != nil {
			return kit.Render(admin.SetupPage("", "", "", "Failed to generate affiliate ID: "+err.Error()))
		}

		affiliate := models.Affiliate{
			AffiliateID:  affiliateID,
			Name:         name,
			Email:        email,
			PasswordHash: string(hash),
			Rate:         0,
			Active:       true,
			ShopURL:      shopURL,
			Balance:      models.NewCurrency(100.00),
		}
		if err := db.Get().Create(&affiliate).Error; err != nil {
			return kit.Render(admin.SetupPage("", "", "", "Failed to create affiliate: "+err.Error()))
		}

		cfg := config.Get()
		cfg.Site.AffiliateID = affiliateID
		if err := config.Save(cfg); err != nil {
			return kit.Render(admin.SetupPage("", "", "", "Failed to save affiliate ID to config: "+err.Error()))
		}
	}

	return kit.Render(admin.SetupPage(email, password, shopURL, ""))
}
