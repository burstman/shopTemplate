package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
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
}

type NotificationConfig struct {
	AdminEmailRecipient string `json:"admin_email_recipient"`
	TelegramBotToken    string `json:"telegram_bot_token"`
	TelegramChatID      string `json:"telegram_chat_id"`
}

type FacebookPixelConfig struct {
	PixelID            string `json:"pixel_id"`
	Currency           string `json:"currency"`
	TrackPurchaseValue bool   `json:"track_purchase_value"`
}

type SiteConfig struct {
	Name          string `json:"name"`
	SupportEmail  string `json:"support_email"`
	NameBgColor   string `json:"name_bg_color"`
	NameTextColor string `json:"name_text_color"`
	Logo          string `json:"logo"`
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
		{Title: "Site Settings", Link: "/admin/site", Icon: "settings"},
		{Title: "Hero Section", Link: "/admin/hero", Icon: "image"},
		{Title: "Homepage Sections", Link: "/admin/sections", Icon: "layout-grid"},
		{Title: "Theme", Link: "/admin/theme", Icon: "palette"},
		{Title: "Notifications", Link: "/admin/notifications", Icon: "bell"},
		{Title: "Facebook Pixel", Link: "/admin/facebook_pixel", Icon: "facebook"},
		{Title: "Categories", Link: "/admin/categories", Icon: "folder-tree"},
		{Title: "Products", Link: "/admin/products", Icon: "shopping-bag"},
		{Title: "Orders", Link: "/admin/orders", Icon: "clipboard-list"},
	}
}

func defaultConfig() *Config {
	return &Config{
		Site: SiteConfig{
			Name:          "Botanica",
			SupportEmail:  "support@botanica.com",
			NameBgColor:   "#2E7D32",
			NameTextColor: "#FFFFFF",
			Logo:          "",
		},
		Notification: NotificationConfig{
			AdminEmailRecipient: "admin@botanica.com",
			TelegramBotToken:    "",
			TelegramChatID:      "",
		},
		FacebookPixel: FacebookPixelConfig{
			PixelID:            "",
			Currency:           "TND",
			TrackPurchaseValue: true,
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
			Copyright: fmt.Sprintf("© %d Botanica. All rights reserved.", time.Now().Year()),
			SocialLinks: []SocialLink{
				{Platform: "Twitter", URL: "https://twitter.com", Icon: "twitter"},
				{Platform: "Facebook", URL: "https://facebook.com", Icon: "facebook"},
				{Platform: "Instagram", URL: "https://instagram.com", Icon: "instagram"},
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
						{Title: "Contact Us", Link: "/contact"},
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
	}
}

// Get loads the configuration from config.json, creating it with defaults if it doesn't exist.
// It uses a singleton pattern to only read from disk once.
func Get() *Config {
	once.Do(func() {
		mu.Lock()
		defer mu.Unlock()
		file, err := os.ReadFile(configFilePath)
		if err != nil {
			if os.IsNotExist(err) {
				cfg = defaultConfig()
				// Use an internal save function which doesn't lock, to avoid deadlock.
				if err := save(cfg); err != nil {
					panic("failed to create config file: " + err.Error())
				}
			} else {
				panic("failed to read config file: " + err.Error())
			}
		} else {
			if err := json.Unmarshal(file, &cfg); err != nil {
				panic("failed to parse config file: " + err.Error())
			}

			// Backfill storefront sidebar if it's missing
			if len(cfg.StorefrontSidebar) == 0 {
				cfg.StorefrontSidebar = defaultConfig().StorefrontSidebar
			}
			// Backfill footer if it's missing
			if cfg.Footer.Copyright == "" {
				cfg.Footer = defaultConfig().Footer
			}
		}
	})
	return cfg
}

// save is an internal, non-locking function to write the config file.
// It should only be called from a function that already holds the mutex.
func save(c *Config) error {
	if err := os.MkdirAll(filepath.Dir(configFilePath), 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(configFilePath, data, 0644)
}

// Save writes the configuration to config.json, acquiring a lock to prevent race conditions.
func Save(newCfg *Config) error {
	mu.Lock()
	defer mu.Unlock()
	cfg = newCfg // Update the singleton instance
	return save(newCfg)
}
