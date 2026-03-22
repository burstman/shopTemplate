package handlers

import (
	"shopTemplate/app/config"
	"shopTemplate/app/db"
	"shopTemplate/app/models"
	"shopTemplate/app/views/components"
	"shopTemplate/app/views/landing"

	"github.com/anthdm/superkit/kit"
)

func HandleLandingIndex(kit *kit.Kit) error {
	// 1. Fetch all settings
	cfg := config.Get()

	// 2. Fetch the specific products selected for the category section
	var categoryProducts []models.Product
	var featuredSection *config.SectionConfig
	for i := range cfg.Sections {
		if cfg.Sections[i].Type == "featured_products" {
			featuredSection = &cfg.Sections[i]
			break
		}
	}

	if featuredSection != nil && featuredSection.Enabled && len(featuredSection.ProductIDs) > 0 {
		var validIDs []string
		for _, id := range featuredSection.ProductIDs {
			if id != "" {
				validIDs = append(validIDs, id)
			}
		}
		if len(validIDs) > 0 {
			if err := db.Get().Where("id IN ?", validIDs).Preload("Categories").Find(&categoryProducts).Error; err != nil {
				return err
			}
		}
	}

	// 3. Fetch all products for the main shop display, ordered by newest first
	var allProducts []models.Product
	if err := db.Get().Order("created_at desc").Preload("Categories").Find(&allProducts).Error; err != nil {
		return err
	}

	// 4. Prepare Carousel Items
	var carouselItems []components.CarouselItem
	if cfg.Hero.Enabled && len(cfg.Hero.Slides) > 0 {
		for _, slide := range cfg.Hero.Slides {
			title := slide.Title
			if title == "" {
				title = cfg.Hero.Title
			}
			subtitle := slide.Subtitle
			if subtitle == "" {
				subtitle = cfg.Hero.Subtitle
			}
			btnText := slide.ButtonText
			if btnText == "" {
				btnText = cfg.Hero.ButtonText
			}
			btnLink := slide.ButtonLink
			if btnLink == "" {
				btnLink = cfg.Hero.ButtonLink
			}
			carouselItems = append(carouselItems, components.CarouselItem{
				Image:       slide.Image,
				Title:       title,
				Description: subtitle,
				ButtonText:  btnText,
				ButtonLink:  btnLink,
			})
		}
	}

	return RenderWithLayout(kit, landing.Index(cfg, categoryProducts, allProducts, carouselItems))
}
