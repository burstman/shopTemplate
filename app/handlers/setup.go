package handlers

import (
	"bytes"
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
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

		// Register with the dashboard
		dashboardURL := os.Getenv("DASHBOARD_URL")
		regSecret := os.Getenv("REGISTRATION_SECRET")
		if dashboardURL != "" && regSecret != "" {
			regPayload := map[string]any{
				"affiliate_id": affiliateID,
				"name":         name,
				"email":        email,
				"shop_url":     shopURL,
				"rate":         0,
			}
			body, _ := json.Marshal(regPayload)
			req, _ := http.NewRequest("POST", dashboardURL+"/api/affiliates/register", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+regSecret)
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				return kit.Render(admin.SetupPage("", "", "", "Failed to register with dashboard: "+err.Error()))
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				respBody, _ := io.ReadAll(resp.Body)
				return kit.Render(admin.SetupPage("", "", "", fmt.Sprintf("Dashboard registration failed (%d): %s", resp.StatusCode, string(respBody))))
			}
			var regResp struct {
				APIKey       string `json:"api_key"`
				DashboardURL string `json:"dashboard_url"`
			}
			if err := json.NewDecoder(resp.Body).Decode(&regResp); err != nil {
				return kit.Render(admin.SetupPage("", "", "", "Failed to decode dashboard response: "+err.Error()))
			}
			db.Get().Model(&models.Affiliate{}).Where("affiliate_id = ?", affiliateID).Updates(map[string]any{
				"api_key":       regResp.APIKey,
				"dashboard_url": regResp.DashboardURL,
			})
		}
	}

	return kit.Render(admin.SetupPage(email, password, shopURL, ""))
}
