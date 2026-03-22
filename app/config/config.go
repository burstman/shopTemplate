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
	Site              SiteConfig      `json:"site"`
	Hero              HeroConfig      `json:"hero"`
	Sections          []SectionConfig `json:"sections"`
	Theme             ThemeConfig     `json:"theme"`
	StorefrontSidebar []MenuItem      `json:"storefront_sidebar"`
	Footer            FooterConfig    `json:"footer"`
}

type SiteConfig struct {
	Name         string `json:"name"`
	SupportEmail string `json:"support_email"`
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
}

type SectionConfig struct {
	Type       string   `json:"type"`
	Title      string   `json:"title"`
	Limit      int      `json:"limit"`
	Enabled    bool     `json:"enabled"`
	ProductIDs []string `json:"product_ids,omitempty"`
}

type ThemeConfig struct {
	PrimaryColor   string `json:"primary_color"`
	SecondaryColor string `json:"secondary_color"`
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
		{Title: "Categories", Link: "/admin/categories", Icon: "folder-tree"},
		{Title: "Add Product", Link: "/products/new", Icon: "plus-circle"},
	}
}

func defaultConfig() *Config {
	return &Config{
		Site: SiteConfig{
			Name:         "Botanica",
			SupportEmail: "support@botanica.com",
		},
		Hero: HeroConfig{
			Enabled:    true,
			Title:      "Nature's Masterpiece",
			Subtitle:   "Discover our indoor plants",
			ButtonText: "Shop Now",
			ButtonLink: "/shop",
			Slides:     []HeroSlide{},
		},
		Sections: []SectionConfig{
			{
				Type:       "featured_products",
				Title:      "Featured Collection",
				Limit:      4,
				Enabled:    true,
				ProductIDs: []string{},
			},
			{
				Type:    "best_sellers",
				Title:   "Best Sellers",
				Limit:   4,
				Enabled: true,
			},
			{
				Type:    "new_arrivals",
				Title:   "New Arrivals",
				Limit:   4,
				Enabled: true,
			},
		},
		Theme: ThemeConfig{
			PrimaryColor:   "#2E7D32",
			SecondaryColor: "#F5F5F5",
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
