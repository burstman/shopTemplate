package handlers

import (
	"crypto/rand"
	"database/sql"
	"fmt"
	"log/slog"
	"math/big"
	"net/http"
	"net/smtp"
	"os"
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

func sendAdminPasswordEmail(to, password, siteName string) bool {
	from := os.Getenv("SMTP_FROM")
	host := os.Getenv("SMTP_HOST")
	port := os.Getenv("SMTP_PORT")
	username := os.Getenv("SMTP_USER")
	smtpPass := os.Getenv("SMTP_PASS")
	if from == "" || host == "" || port == "" {
		return false
	}
	auth := smtp.PlainAuth("", username, smtpPass, host)
	header := fmt.Sprintf("From: %s\r\n", from)
	header += fmt.Sprintf("To: %s\r\n", to)
	header += fmt.Sprintf("Subject: Admin Account Created - %s\r\n", siteName)
	header += "MIME-version: 1.0\r\n"
	header += "Content-Type: text/plain; charset=\"UTF-8\"\r\n\r\n"
	body := fmt.Sprintf(
		"Hello,\n\n"+
			"Your admin account has been created for %s.\n\n"+
			"Email: %s\n"+
			"Password: %s\n\n"+
			"Please log in and change your password as soon as possible.\n\n"+
			"Thank you,\n%s",
		siteName, to, password, siteName,
	)
	if err := smtp.SendMail(host+":"+port, auth, from, []string{to}, []byte(header+body)); err != nil {
		slog.Error("failed to send admin password email", "err", err)
		return false
	}
	return true
}

func HandleSetupIndex(kit *kit.Kit) error {
	aff := config.AffiliateFromContext(kit.Request.Context())
	slog.Info("setup index", "host", kit.Request.Host, "affiliate_found", aff != nil, "has_password", aff != nil && aff.PasswordHash != "")
	if aff != nil && aff.PasswordHash != "" {
		return kit.Redirect(http.StatusSeeOther, "/login")
	}
	return kit.Render(admin.SetupPage("", "", "", "", false))
}

func HandleSetupCreate(kit *kit.Kit) error {
	aff := config.AffiliateFromContext(kit.Request.Context())
	slog.Info("setup create", "host", kit.Request.Host, "affiliate_found", aff != nil, "affiliate_id", func() string { if aff != nil { return aff.AffiliateID }; return "" }())

	name := kit.Request.FormValue("name")
	email := kit.Request.FormValue("email")
	slog.Info("setup create form", "name", name, "email", email)

	if name == "" || email == "" {
		slog.Warn("setup create validation failed", "reason", "missing fields")
		return kit.Render(admin.SetupPage("", "", "", "All fields are required.", false))
	}

	// Check if email is already used by another affiliate
	var existingAff models.Affiliate
	excludeID := ""
	if aff != nil {
		excludeID = aff.AffiliateID
	}
	if err := db.Get().Where("email = ? AND affiliate_id != ?", email, excludeID).First(&existingAff).Error; err == nil {
		slog.Warn("setup create email conflict", "email", email, "existing_id", existingAff.AffiliateID)
		return kit.Render(admin.SetupPage("", "", "", "This email is already registered to another shop. Each shop must use a unique email.", false))
	}

	// Check if the email is authorized
	var authorizedEmails []string
	db.Get().Model(&models.Affiliate{}).Where("authorized_email <> ''").Pluck("authorized_email", &authorizedEmails)
	if len(authorizedEmails) > 0 {
		slog.Info("setup create authorized check", "authorized_list", authorizedEmails)
		authorized := false
		for _, ae := range authorizedEmails {
			if ae == email {
				authorized = true
				break
			}
		}
		if !authorized {
			slog.Warn("setup create unauthorized email", "email", email)
			return kit.Render(admin.SetupPage("", "", "", "Your email is not authorized for setup. Contact the shop owner.", false))
		}
	}

	scheme := "https"
	if kit.Request.TLS == nil {
		scheme = "http"
	}
	shopURL := fmt.Sprintf("%s://%s", scheme, kit.Request.Host)
	slog.Info("setup create shop_url", "shop_url", shopURL)

	password, err := generateSetupPassword(12)
	if err != nil {
		slog.Error("setup create password generation failed", "err", err)
		return err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		slog.Error("setup create bcrypt failed", "err", err)
		return err
	}

	// Update or create affiliate for this host
	if aff == nil {
		affiliateID, err := generateAffiliateID()
		if err != nil {
			slog.Error("setup create affiliate ID generation failed", "err", err)
			return kit.Render(admin.SetupPage("", "", "", "Failed to generate affiliate ID: "+err.Error(), false))
		}
		aff = &models.Affiliate{
			AffiliateID:  affiliateID,
			Name:         name,
			Email:        email,
			PasswordHash: string(hash),
			Rate:         0,
			Active:       true,
			ShopURL:      shopURL,
			Balance:      models.NewCurrency(100.00),
		}
		slog.Info("setup create creating new affiliate", "affiliate_id", affiliateID, "shop_url", shopURL)
		if err := db.Get().Create(aff).Error; err != nil {
			slog.Error("setup create affiliate creation failed", "err", err)
			return kit.Render(admin.SetupPage("", "", "", "Failed to create affiliate: "+err.Error(), false))
		}
	} else {
		slog.Info("setup create updating existing affiliate", "affiliate_id", aff.AffiliateID)
		db.Get().Model(aff).Updates(map[string]interface{}{
			"name":          name,
			"email":         email,
			"password_hash": string(hash),
		})
	}

	// Create or reuse user for this shop
	var existing models.User
	if err := db.Get().Where("email = ?", email).First(&existing).Error; err != nil {
		user := models.User{
			Email:           email,
			FirstName:       name,
			PasswordHash:    string(hash),
			Role:            "admin",
			EmailVerifiedAt: sql.NullTime{Time: time.Now(), Valid: true},
			AffiliateID:     aff.AffiliateID,
		}
		slog.Info("setup create creating new user", "email", email)
		if err := db.Get().Create(&user).Error; err != nil {
			slog.Error("setup create user creation failed", "err", err)
			return kit.Render(admin.SetupPage("", "", "", "Failed to create admin: "+err.Error(), false))
		}
	} else {
		slog.Info("setup create updating existing user", "email", email, "user_id", existing.ID)
		db.Get().Model(&existing).Update("password_hash", string(hash))
	}

	cfg := config.Get()
	cfg.Site.AffiliateID = aff.AffiliateID
	slog.Info("setup create saving config", "affiliate_id", aff.AffiliateID)
	if err := config.Save(cfg); err != nil {
		slog.Error("setup create config save failed", "err", err)
		return kit.Render(admin.SetupPage("", "", "", "Failed to save affiliate ID to config: "+err.Error(), false))
	}

	// Send password via email if SMTP is configured
	emailSent := sendAdminPasswordEmail(email, password, cfg.Site.Name)
	slog.Info("setup create complete", "email_sent", emailSent)

	return kit.Render(admin.SetupPage(email, password, shopURL, "", emailSent))
}
