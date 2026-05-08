package config

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"sync"
	"strings"
	"time"

	"shopTemplate/app/db"
	"shopTemplate/app/models"
)

const configFilePath = "app/config/config.json"

var (
	cfg  *Config
	once sync.Once
	mu   sync.Mutex
)

type Config struct {
	Site              SiteConfig          `json:"site"`
	Notification      NotificationConfig  `json:"notification"`
	FacebookPixel     FacebookPixelConfig `json:"facebook_pixel"`
	Hero              HeroConfig          `json:"hero"`
	Sections          []SectionConfig     `json:"sections"`
	Theme             ThemeConfig         `json:"theme"`
	StorefrontSidebar []MenuItem          `json:"storefront_sidebar"`
	Footer            FooterConfig        `json:"footer"`
	Payment           PaymentConfig       `json:"payment"`
	Chat              ChatConfig          `json:"chat"`
}

type NotificationConfig struct {
	AdminEmailRecipient string `json:"admin_email_recipient"`
	TelegramBotToken    string `json:"telegram_bot_token"`
	TelegramChatID      string `json:"telegram_chat_id"`
}

type FacebookPixelConfig struct {
	PixelID            string `json:"pixel_id"`
	TrackPurchaseValue bool   `json:"track_purchase_value"`
	AccessToken        string `json:"access_token"`
	DomainVerification string `json:"domain_verification"`
	TestEventCode      string `json:"test_event_code"`
}



type SiteConfig struct {
	Name          string          `json:"name"`
	SupportEmail  string          `json:"support_email"`
	NameBgColor   string          `json:"name_bg_color"`
	NameTextColor string          `json:"name_text_color"`
	Logo          string          `json:"logo"`
	Currency      string          `json:"currency"`
	ShowQuickView bool            `json:"show_quick_view"`
	ShowOrderNow  bool            `json:"show_order_now"`
	ShowAddToCart bool            `json:"show_add_to_cart"`
	Bundles       []models.Bundle `json:"bundles"`
}

type HeroConfig struct {
	Enabled    bool        `json:"enabled"`
	Title      string      `json:"title"`
	Subtitle   string      `json:"subtitle"`
	ButtonText string      `json:"button_text"`
	ButtonLink string      `json:"button_link"`
	Slides     []HeroSlide `json:"slides"`
}

type HeroSlide struct {
	Image      string `json:"image"`
	Title      string `json:"title,omitempty"`
	Subtitle   string `json:"subtitle,omitempty"`
	ButtonText string `json:"button_text,omitempty"`
	ButtonLink string `json:"button_link,omitempty"`
	ProductID  *uint  `json:"product_id,omitempty"`
}

type CategorySectionItem struct {
	Title      string `json:"title"`
	CategoryID string `json:"category_id"`
	Image      string `json:"image"`
}

type SectionConfig struct {
	Type           string                `json:"type"`
	Title          string                `json:"title"`
	Limit          int                   `json:"limit"`
	Enabled        bool                  `json:"enabled"`
	ProductIDs     []string              `json:"product_ids,omitempty"`
	CategoryID     string                `json:"category_id,omitempty"`
	Image          string                `json:"image,omitempty"`
	CategoryItems  []CategorySectionItem `json:"category_items,omitempty"`
	TitleBgColor   string                `json:"title_bg_color"`
	TitleTextColor string                `json:"title_text_color"`
}

type ThemeConfig struct {
	PrimaryColor             string `json:"primary_color"`
	SecondaryColor           string `json:"secondary_color"`
	HeaderBgColor            string `json:"header_bg_color"`
	HeaderBgOpacity          int    `json:"header_bg_opacity"`
	PageBgColor              string `json:"page_bg_color"`
	FooterBgColor            string `json:"footer_bg_color"`
	ContentBgGradientEnabled bool   `json:"content_bg_gradient_enabled"`
	ContentBgGradientStart   string `json:"content_bg_gradient_start"`
	ContentBgGradientEnd     string `json:"content_bg_gradient_end"`
	ContentBgGradientAngle   int    `json:"content_bg_gradient_angle"`
	ContentBgColorEnabled    bool   `json:"content_bg_color_enabled"`
	ContentBgColor           string `json:"content_bg_color"`
}

type FooterConfig struct {
	Copyright   string       `json:"copyright"`
	SocialLinks []SocialLink `json:"social_links"`
	LinkColumns []LinkColumn `json:"link_columns"`
}

type ChatConfig struct {
	PrimaryColor      string `json:"primary_color"`
	HeaderTextColor   string `json:"header_text_color"`
	ClientBubbleColor string `json:"client_bubble_color"`
	ClientTextColor   string `json:"client_text_color"`
	AdminBubbleColor  string `json:"admin_bubble_color"`
	AdminTextColor    string `json:"admin_text_color"`
	EnablePopup       bool   `json:"enable_popup"`
	PopupTimeout      int    `json:"popup_timeout"`
}

type PaymentConfig struct {
	EnableCOD       bool   `json:"enable_cod"`
	EnableFlouci    bool   `json:"enable_flouci"`
	FlouciPublicKey string `json:"flouci_public_key"`
	FlouciAppToken  string `json:"flouci_app_token"`
}

type SocialLink struct {
	Platform string `json:"platform"`
	URL      string `json:"url"`
	Icon     string `json:"icon"`
}

type LinkColumn struct {
	Title string     `json:"title"`
	Links []MenuItem `json:"links"`
}

type MenuItem struct {
	Title string `json:"title"`
	Link  string `json:"link"`
	Icon  string `json:"icon,omitempty"`
}

func GetAdminSidebar() []MenuItem {
	return []MenuItem{
		{Title: "Back to Store", Link: "/", Icon: "arrow-left"},
		{Title: "Dashboard", Link: "/admin/dashboard", Icon: "layout-dashboard"},
		{Title: "Site Settings", Link: "/admin/site", Icon: "settings"},
		{Title: "Hero Section", Link: "/admin/hero", Icon: "image"},
		{Title: "Homepage Sections", Link: "/admin/sections", Icon: "layout-grid"},
		{Title: "Theme", Link: "/admin/theme", Icon: "palette"},
		{Title: "Notifications", Link: "/admin/notifications", Icon: "bell"},
		{Title: "Facebook Pixel", Link: "/admin/facebook_pixel", Icon: "facebook"},
		{Title: "Categories", Link: "/admin/categories", Icon: "folder-tree"},
		{Title: "Products", Link: "/admin/products", Icon: "shopping-bag"},
		{Title: "Orders", Link: "/admin/orders", Icon: "clipboard-list"},
		{Title: "Payment Methods", Link: "/admin/payment", Icon: "credit-card"},
		{Title: "Chat Settings", Link: "/admin/chat_settings", Icon: "message-square"},
		{Title: "Social Links", Link: "/admin/social_links", Icon: "share-2"},
	}
}

func defaultConfig() *Config {
	return &Config{
		Site: SiteConfig{
			Name:          "BEST SHOP",
			SupportEmail:  "support@bestshop.com",
			NameBgColor:   "#2E7D32",
			NameTextColor: "#FFFFFF",
			Logo:          "",
			Currency:      "TND",
			ShowQuickView: true,
			ShowOrderNow:  true,
			ShowAddToCart: true,
			Bundles: []models.Bundle{
				{Quantity: 2, DiscountPercentage: 10},
				{Quantity: 3, DiscountPercentage: 15},
			},
		},
		Notification: NotificationConfig{
			AdminEmailRecipient: "admin@bestshop.com",
			TelegramBotToken:    "",
			TelegramChatID:      "",
		},
		FacebookPixel: FacebookPixelConfig{
			PixelID:            "",
			TrackPurchaseValue: true,
			AccessToken:        "",
			DomainVerification: "",
			TestEventCode:      "",
		},
		Hero: HeroConfig{
			Enabled:    true,
			Title:      "Nature's Masterpiece",
			Subtitle:   "Discover our indoor plants",
			ButtonText: "Shop Now",
			ButtonLink: "/",
			Slides:     []HeroSlide{},
		},
		Sections: []SectionConfig{
			{
				Type:           "featured_products",
				Title:          "Featured Collection",
				Limit:          4,
				Enabled:        true,
				ProductIDs:     []string{},
				TitleBgColor:   "#2E7D32",
				TitleTextColor: "#FFFFFF",
			},
			{
				Type:           "best_sellers",
				Title:          "Best Sellers",
				Limit:          4,
				Enabled:        true,
				ProductIDs:     []string{},
				TitleBgColor:   "#2E7D32",
				TitleTextColor: "#FFFFFF",
			},
			{
				Type:           "new_arrivals",
				Title:          "New Arrivals",
				Limit:          4,
				Enabled:        true,
				ProductIDs:     []string{},
				TitleBgColor:   "#2E7D32",
				TitleTextColor: "#FFFFFF",
			},
		},
		Theme: ThemeConfig{
			PrimaryColor:             "#2E7D32",
			SecondaryColor:           "#F5F5F5",
			HeaderBgColor:            "#FFFFFF",
			HeaderBgOpacity:          100,
			PageBgColor:              "#FFFFFF",
			FooterBgColor:            "#F9FAFB",
			ContentBgGradientEnabled: false,
			ContentBgGradientStart:   "#FFFFFF",
			ContentBgGradientEnd:     "#F3F4F6",
			ContentBgGradientAngle:   180,
			ContentBgColorEnabled:    false,
			ContentBgColor:           "#FFFFFF",
		},
		StorefrontSidebar: []MenuItem{},
		Footer: FooterConfig{
			Copyright: fmt.Sprintf("© %d BEST SHOP. All rights reserved.", time.Now().Year()),
			SocialLinks: []SocialLink{
				{Platform: "Facebook", URL: "https://facebook.com", Icon: "facebook"},
				{Platform: "Instagram", URL: "https://instagram.com", Icon: "instagram"},
				{Platform: "TikTok", URL: "https://tiktok.com", Icon: "tiktok"},
				{Platform: "WhatsApp", URL: "https://wa.me/21600000000", Icon: "whatsapp"},
			},
			LinkColumns: []LinkColumn{
				{
					Title: "Shop",
					Links: []MenuItem{
						{Title: "Products", Link: "/products"},
						{Title: "Best Sellers", Link: "/best-sellers"},
						{Title: "New Arrivals", Link: "/new-arrivals"},
					},
				},
				{
					Title: "About",
					Links: []MenuItem{
						{Title: "Our Story", Link: "/about"},
						{Title: "FAQs", Link: "/faq"},
					},
				},
				{
					Title: "Legal",
					Links: []MenuItem{
						{Title: "Privacy Policy", Link: "/privacy"},
						{Title: "Terms of Service", Link: "/terms"},
					},
				},
			},
		},
		Chat: ChatConfig{
			PrimaryColor:      "#2E7D32",
			HeaderTextColor:   "#FFFFFF",
			ClientBubbleColor: "#2E7D32",
			ClientTextColor:   "#FFFFFF",
			AdminBubbleColor:  "#FFFFFF",
			AdminTextColor:    "#1F2937",
			EnablePopup:       true,
			PopupTimeout:      8,
		},
		Payment: PaymentConfig{
			EnableCOD:       true,
			EnableFlouci:    false,
			FlouciPublicKey: "",
			FlouciAppToken:  "",
		},
	}
}

// Get loads the configuration from config.json, creating it with defaults if it doesn't exist.
// It uses a singleton pattern to only read from disk once.
func Get() *Config {
	once.Do(func() {
		// 1. Try to load from Postgres settings table
		var setting models.Setting
		err := db.Get().Where("key = ?", "app_config").First(&setting).Error
		if err == nil && setting.Value != "" {
			if err := json.Unmarshal([]byte(setting.Value), &cfg); err == nil {
				slog.Info("configuration loaded from postgres")
				backfill(cfg)
				return
			}
		}

		// 2. Migration Fallback: Try to read from the legacy JSON file
		if file, err := os.ReadFile(configFilePath); err == nil {
			if err := json.Unmarshal(file, &cfg); err == nil {
				slog.Info("configuration migrated from local file to postgres")
				backfill(cfg)
				save(cfg) // Persist legacy file to DB
				return
			}
		}

		// 3. Final Fallback: Use hardcoded defaults
		slog.Warn("no configuration found in DB or file, using hardcoded defaults")
		cfg = defaultConfig()
		save(cfg)
	})
	return cfg
}

func backfill(c *Config) {
	if len(c.StorefrontSidebar) == 0 {
		c.StorefrontSidebar = defaultConfig().StorefrontSidebar
	}
	if c.Footer.Copyright == "" || strings.Contains(c.Footer.Copyright, "Botanica") {
		c.Footer.Copyright = fmt.Sprintf("© %d %s. All rights reserved.", time.Now().Year(), c.Site.Name)
	}
}

// save is an internal, non-locking function to write the config to the DB.
// It should only be called from a function that already holds the mutex.
func save(c *Config) error {
	data, err := json.Marshal(c)
	if err != nil {
		return err
	}

	// PostgreSQL Upsert logic using ON CONFLICT
	return db.Get().Exec(`
		INSERT INTO settings (key, value, created_at, updated_at) 
		VALUES (?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP) 
		ON CONFLICT (key) 
		DO UPDATE SET value = EXCLUDED.value, updated_at = EXCLUDED.updated_at`,
		"app_config", string(data)).Error
}

// Save writes the configuration to config.json, acquiring a lock to prevent race conditions.
func Save(newCfg *Config) error {
	mu.Lock()
	defer mu.Unlock()
	if err := save(newCfg); err != nil {
		return err
	}
	// Only update the singleton memory instance if the DB save succeeded
	cfg = newCfg 
	return nil
}
