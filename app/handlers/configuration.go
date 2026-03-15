package handlers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"shopTemplate/app/db"
	"shopTemplate/app/models"
	"strconv"
	"strings"

	"shopTemplate/app/views/configuration"

	"github.com/anthdm/superkit/kit"
)

func HandleConfigurationIndex(kit *kit.Kit) error {
	user, ok := kit.Auth().(models.AuthUser)
	if !ok || user.Role != "admin" {
		return kit.Redirect(http.StatusSeeOther, "/")
	}

	defaultSettings := map[string]string{
		"carousel_count":        "0",
		"category_count":        "3",
		"best_seller_count":     "4",
		"favorite_plants_count": "4",
		"category_products":     "",
	}

	for key, value := range defaultSettings {
		db.Get().Where(models.Setting{Key: key}).FirstOrCreate(&models.Setting{Key: key, Value: value})
	}
	// Fetch settings from DB after ensuring they exist
	var settings []models.Setting
	err := db.Get().Find(&settings).Error
	if err != nil {
		return err
	}

	// Convert to map for easy access in view
	configMap := make(map[string]string)
	for _, s := range settings {
		configMap[s.Key] = s.Value
	}

	isHTMX := kit.Request.Header.Get("HX-Request") == "true"
	requestedCount := kit.Request.URL.Query().Get("category_count")

	var selectedIDs []string

	if isHTMX {
		// If HTMX, preserve the current selection from the UI (hx-include)
		selectedIDs = kit.Request.URL.Query()["category_products"]
		if requestedCount != "" {
			configMap["category_count"] = requestedCount
		}
	} else {
		// Initial load: fallback to DB
		if val := configMap["category_products"]; val != "" {
			selectedIDs = strings.Split(val, ",")
		}
	}

	// Update map so template shows correct selections
	configMap["category_products"] = strings.Join(selectedIDs, ",")

	var productsForSelector []models.Product
	db.Get().Order("category, name").Find(&productsForSelector)

	// If this is an HTMX request for the selector, render only that component
	if isHTMX {
		return kit.Render(configuration.CategoryProductSelector(productsForSelector, configMap["category_products"], configMap["category_count"]))
	}

	return RenderWithLayout(kit, configuration.Index(configMap, productsForSelector))
}

func HandleConfigurationUpdate(kit *kit.Kit) error {
	user, ok := kit.Auth().(models.AuthUser)
	if !ok || user.Role != "admin" {
		return kit.Redirect(http.StatusSeeOther, "/")
	}

	// Parse form values
	err := kit.Request.ParseMultipartForm(32 << 20) // 32MB max
	if err != nil {
		return err
	}

	// Handle Carousel Image Uploads
	files := kit.Request.MultipartForm.File["carousel_images"]
	if len(files) > 0 {
		carouselPath := "public/images/carousel"
		// Create directory if not exists
		if err := os.MkdirAll(carouselPath, 0755); err != nil {
			return err
		}

		for i, fileHeader := range files {
			file, err := fileHeader.Open()
			if err != nil {
				return err
			}
			defer file.Close()

			// Auto-rename to slide_1.ext, slide_2.ext, etc.
			ext := filepath.Ext(fileHeader.Filename)
			dstName := fmt.Sprintf("slide_%d%s", i+1, ext)
			dstPath := filepath.Join(carouselPath, dstName)

			dst, err := os.Create(dstPath)
			if err != nil {
				return err
			}
			defer dst.Close()

			if _, err := io.Copy(dst, file); err != nil {
				return err
			}
		}

		// Update carousel_count setting
		err = db.Get().Where(models.Setting{Key: "carousel_count"}).
			Assign(models.Setting{Value: strconv.Itoa(len(files))}).
			FirstOrCreate(&models.Setting{}).Error
		if err != nil {
			return err
		}
	}

	// Handle category products selection
	selectedIDs := kit.Request.MultipartForm.Value["category_products"]
	productsValue := strings.Join(selectedIDs, ",")
	err = db.Get().Where(models.Setting{Key: "category_products"}).
		Assign(models.Setting{Value: productsValue}).
		FirstOrCreate(&models.Setting{}).Error
	if err != nil {
		return err
	}

	// Iterate over posted values and update/create settings
	for key, values := range kit.Request.MultipartForm.Value {
		if len(values) > 0 && key != "carousel_images" && key != "category_products" {
			// Upsert logic: Save key-value pair
			err = db.Get().Where(models.Setting{Key: key}).Assign(models.Setting{Value: values[0]}).FirstOrCreate(&models.Setting{}).Error
			if err != nil {
				return err
			}
		}
	}

	return kit.Redirect(http.StatusSeeOther, "/configuration")
}
