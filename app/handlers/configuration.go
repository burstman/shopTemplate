package handlers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"shopTemplate/app/config"
	"shopTemplate/app/db"
	"shopTemplate/app/models"
	"shopTemplate/app/views/layouts"
	"strconv"
	"time"

	"shopTemplate/app/views/configuration"

	"github.com/anthdm/superkit/kit"
	"github.com/go-chi/chi/v5"
)

func HandleAdminSettings(kit *kit.Kit) error {
	user, ok := kit.Auth().(models.AuthUser)
	if !ok || user.Role != "admin" {
		return kit.Redirect(http.StatusSeeOther, "/")
	}

	section := chi.URLParam(kit.Request, "section")
	if section == "" {
		section = "site"
	}

	cfg := config.Get()

	var productsForSelector []models.Product
	if section == "sections" {
		db.Get().Order("name").Find(&productsForSelector)
	}

	activePath := "/admin/" + section
	sidebar := config.GetAdminSidebar()
	content := configuration.Index(cfg, productsForSelector, section)
	return RenderWithLayout(kit, layouts.AdminPage(sidebar, activePath, content))
}

func HandleAdminSettingsUpdate(kit *kit.Kit) error {
	user, ok := kit.Auth().(models.AuthUser)
	if !ok || user.Role != "admin" {
		return kit.Redirect(http.StatusSeeOther, "/")
	}

	section := chi.URLParam(kit.Request, "section")
	if err := kit.Request.ParseMultipartForm(32 << 20); err != nil { // 32MB max
		return err
	}

	cfg := config.Get()

	switch section {
	case "site":
		cfg.Site.Name = kit.Request.FormValue("site_name")
		cfg.Site.SupportEmail = kit.Request.FormValue("support_email")

	case "hero":
		cfg.Hero.Enabled = kit.Request.FormValue("hero_enabled") == "on"
		cfg.Hero.Title = kit.Request.FormValue("hero_title")
		cfg.Hero.Subtitle = kit.Request.FormValue("hero_subtitle")
		cfg.Hero.ButtonText = kit.Request.FormValue("hero_button_text")
		cfg.Hero.ButtonLink = kit.Request.FormValue("hero_button_link")

		// Process Existing Slides
		var updatedSlides []config.HeroSlide
		for i, slide := range cfg.Hero.Slides {
			prefix := fmt.Sprintf("slide_%d_", i)
			if kit.Request.FormValue(prefix+"delete") == "on" {
				if len(slide.Image) > 0 && slide.Image[0] == '/' {
					os.Remove(slide.Image[1:])
				}
				continue
			}
			slide.Title = kit.Request.FormValue(prefix + "title")
			slide.Subtitle = kit.Request.FormValue(prefix + "subtitle")
			slide.ButtonText = kit.Request.FormValue(prefix + "button_text")
			slide.ButtonLink = kit.Request.FormValue(prefix + "button_link")
			updatedSlides = append(updatedSlides, slide)
		}

		// Handle New Uploads
		files := kit.Request.MultipartForm.File["carousel_images"]
		if len(files) > 0 {
			carouselPath := "public/images/carousel"
			os.MkdirAll(carouselPath, 0755)
			for _, fileHeader := range files {
				file, err := fileHeader.Open()
				if err != nil {
					continue
				}
				defer file.Close()
				ext := filepath.Ext(fileHeader.Filename)
				dstName := fmt.Sprintf("slide_%d%s", time.Now().UnixNano(), ext)
				dstPath := filepath.Join(carouselPath, dstName)
				dst, err := os.Create(dstPath)
				if err != nil {
					continue
				}
				defer dst.Close()
				io.Copy(dst, file)
				updatedSlides = append(updatedSlides, config.HeroSlide{Image: "/" + dstPath})
			}
		}
		cfg.Hero.Slides = updatedSlides

	case "sections":
		for i := range cfg.Sections {
			secType := cfg.Sections[i].Type
			cfg.Sections[i].Enabled = kit.Request.FormValue(secType+"_enabled") == "on"
			cfg.Sections[i].Title = kit.Request.FormValue(secType + "_title")
			if limit, err := strconv.Atoi(kit.Request.FormValue(secType + "_limit")); err == nil {
				cfg.Sections[i].Limit = limit
			}
			if secType == "featured_products" {
				cfg.Sections[i].ProductIDs = kit.Request.MultipartForm.Value["featured_products_product_ids"]
			}
		}

	case "theme":
		cfg.Theme.PrimaryColor = kit.Request.FormValue("theme_primary_color")
		cfg.Theme.SecondaryColor = kit.Request.FormValue("theme_secondary_color")
	}

	config.Save(cfg)

	return kit.Redirect(http.StatusSeeOther, "/admin/"+section)
}

func HandleConfigurationIndex(kit *kit.Kit) error {
	return kit.Redirect(http.StatusSeeOther, "/admin/site")
}

func HandleConfigurationUpdate(kit *kit.Kit) error {
	return kit.Redirect(http.StatusSeeOther, "/admin/site")
}
