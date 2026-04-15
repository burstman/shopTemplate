package handlers

import (
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"shopTemplate/app/config"
	"shopTemplate/app/db"
	"shopTemplate/app/helpers"
	"shopTemplate/app/models"
	viewerrors "shopTemplate/app/views/errors"
	"shopTemplate/app/views/layouts"
	"shopTemplate/app/views/products"

	"github.com/anthdm/superkit/kit"
	"github.com/anthdm/superkit/validate"
	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
)

func HandleAdminProductsIndex(kit *kit.Kit) error {
	user, ok := kit.Auth().(models.AuthUser)
	if !ok || user.Role != "admin" {
		return kit.Redirect(http.StatusSeeOther, "/")
	}

	pageStr := kit.Request.URL.Query().Get("page")
	page, _ := strconv.Atoi(pageStr)
	if page < 1 {
		page = 1
	}
	perPage := 10

	var total int64
	err := db.Get().Model(&models.Product{}).Count(&total).Error
	if err != nil {
		return err
	}
	totalPages := int(math.Ceil(float64(total) / float64(perPage)))
	offset := (page - 1) * perPage

	var productsList []models.Product
	db.Get().Preload("Categories").Order("created_at desc").Limit(perPage).Offset(offset).Find(&productsList)

	activePath := "/admin/products"
	sidebar := config.GetAdminSidebar()
	// successMsg is no longer used as toast messages are removed
	content := products.AdminList(productsList, page, totalPages)
	return RenderWithLayout(kit, layouts.AdminPage(sidebar, activePath, content))
}

// HandleProductsIndex renders the public product listing page.
func HandleProductsIndex(kit *kit.Kit) error {
	pageStr := kit.Request.URL.Query().Get("page")
	page, _ := strconv.Atoi(pageStr)
	if page < 1 {
		page = 1
	}
	perPage := 12

	var total int64
	err := db.Get().Model(&models.Product{}).Count(&total).Error
	if err != nil {
		return err
	}
	totalPages := int(math.Ceil(float64(total) / float64(perPage)))
	offset := (page - 1) * perPage

	var productsList []models.Product
	if err := db.Get().Preload("Categories").Order("created_at desc").Limit(perPage).Offset(offset).Find(&productsList).Error; err != nil {
		return err
	}

	cfg := config.Get()
	return RenderWithLayout(kit, products.Index(productsList, page, totalPages, cfg))
}

func HandleProductNew(kit *kit.Kit) error {
	user, ok := kit.Auth().(models.AuthUser)
	if !ok || user.Role != "admin" {
		return kit.Redirect(http.StatusSeeOther, "/")
	}

	categories := helpers.GetCategoryTree()

	activePath := "/products/new"
	sidebar := config.GetAdminSidebar()
	content := products.New(categories, products.CreateForm{
		Errors:     make(validate.Errors),
		Categories: categories,
	})
	return RenderWithLayout(kit, layouts.AdminPage(sidebar, activePath, content))
}

func HandleProductCreate(kit *kit.Kit) error {
	user, ok := kit.Auth().(models.AuthUser)
	if !ok || user.Role != "admin" {
		return kit.Redirect(http.StatusSeeOther, "/")
	}

	// Parse form (max 10MB)
	err := kit.Request.ParseMultipartForm(10 << 20)
	if err != nil {
		return err
	}

	// 1. Handle Image Upload
	errors := make(validate.Errors)
	name := kit.Request.FormValue("name")
	priceStr := kit.Request.FormValue("price")
	description := kit.Request.FormValue("description")
	promotionPriceStr := kit.Request.FormValue("promotion_price")
	stockStr := kit.Request.FormValue("stock")
	file, header, err := kit.Request.FormFile("image")
	categoryIDs := kit.Request.MultipartForm.Value["categories"]
	if name == "" {
		errors.Add("name", "Name is required")
	}
	var price float64
	if priceStr == "" {
		errors.Add("price", "Price is required")
	} else {
		var err error
		price, err = strconv.ParseFloat(priceStr, 64)
		if err != nil {
			errors.Add("price", "Price must be a valid number")
		}
	}
	var promotionPrice float64
	if promotionPriceStr != "" {
		p, err := strconv.ParseFloat(promotionPriceStr, 64)
		if err != nil {
			errors.Add("promotion_price", "Promotion price must be a valid number")
		} else {
			promotionPrice = p
		}
	}
	if err != nil {
		errors.Add("image", "Image is required")
	}
	var stock int
	if stockStr != "" {
		stock, err = strconv.Atoi(stockStr)
		if err != nil {
			errors.Add("stock", "Stock must be a valid integer")
		}
	}

	hasCategory := false
	for _, id := range categoryIDs {
		if id != "" {
			hasCategory = true
			break
		}
	}
	if !hasCategory {
		errors.Add("categories", "Please select at least one category")
	}

	if len(errors) > 0 {
		categories := helpers.GetCategoryTree()
		activePath := "/products/new"
		sidebar := config.GetAdminSidebar()
		form := products.CreateForm{
			Values: map[string]string{
				"name":            name,
				"price":           priceStr,
				"description":     description,
				"promotion_price": promotionPriceStr,
				"stock":           stockStr,
			},
			Categories:          categories,
			SelectedCategoryIDs: categoryIDs,
			Errors:              errors,
		}
		content := products.New(categories, form)
		return RenderWithLayout(kit, layouts.AdminPage(sidebar, activePath, content))
	}

	// Handle Image Upload
	defer file.Close()

	uploadPath := "public/images/products"
	if err := os.MkdirAll(uploadPath, 0755); err != nil {
		return err
	}

	// Auto-rename: plant_{timestamp}.ext
	ext := filepath.Ext(header.Filename)
	newFileName := fmt.Sprintf("plant_%d%s", time.Now().UnixNano(), ext)
	fullPath := filepath.Join(uploadPath, newFileName)

	// Save file
	dst, err := os.Create(fullPath)
	if err != nil {
		return err
	}
	defer dst.Close()
	io.Copy(dst, file)

	// 2. Save Product to DB
	product := models.Product{
		Name:           name,
		Description:    description,
		Price:          price,
		PromotionPrice: promotionPrice,
		Stock:          stock,
		Image:          "/" + fullPath,
	}

	if err := db.Get().Create(&product).Error; err != nil {
		return err
	}

	// Associate categories
	if len(categoryIDs) > 0 {
		var categoriesToAssign []models.Category
		if err := db.Get().Find(&categoriesToAssign, categoryIDs).Error; err == nil {
			db.Get().Model(&product).Association("Categories").Replace(categoriesToAssign)
		}
	}

	// Redirect to refresh the page (or you could return an OOB swap for the grid)
	return kit.Redirect(http.StatusSeeOther, "/admin/products")
}

// HandleProductDeleteConfirm renders a confirmation modal for product deletion.
func HandleProductDeleteConfirm(kit *kit.Kit) error {
	user, ok := kit.Auth().(models.AuthUser)
	if !ok || user.Role != "admin" {
		return kit.Redirect(http.StatusSeeOther, "/")
	}

	idStr := chi.URLParam(kit.Request, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return errors.New("invalid product ID")
	}

	var product models.Product
	if err := db.Get().First(&product, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return kit.Render(viewerrors.Error404())
		}
		return err
	}

	return kit.Render(products.DeleteModal(product))
}

// HandleProductDelete handles the deletion of a product by its ID.
// Only users with the "admin" role are authorized to perform this action.
// The function retrieves the product ID from the URL parameters, validates it,
// and performs a soft delete operation on the product record.
// If the user is unauthorized or the product ID is invalid, appropriate errors are returned.
// On successful deletion, an empty response with HTTP status 200 is returned.
func HandleProductDelete(kit *kit.Kit) error {
	user, ok := kit.Auth().(models.AuthUser)
	if !ok || user.Role != "admin" {
		log.Printf("Unauthorized delete attempt by user: %v", user)
		return kit.Redirect(http.StatusForbidden, "/")
	}

	idStr := chi.URLParam(kit.Request, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return errors.New("invalid product ID")
	}

	var product models.Product
	if err := db.Get().First(&product, id).Error; err != nil {
		return err
	}

	if len(product.Image) > 1 && product.Image[0] == '/' {
		os.Remove(product.Image[1:])
	}

	if err := db.Get().Delete(&product).Error; err != nil {
		return err
	}

	// Toast message removed as per user request.
	// HTMX on the client side should handle removing the element from the DOM.
	return nil
}

// HandleProductEdit handles the editing of a product by an admin user.
// It checks if the authenticated user has the "admin" role, retrieves the product ID from the URL,
// fetches the corresponding product from the database, and renders the product edit modal.
// If the user is not an admin or the product ID is invalid, it redirects to the home page or returns an error.
func HandleProductEdit(kit *kit.Kit) error {
	user, ok := kit.Auth().(models.AuthUser)
	if !ok || user.Role != "admin" {
		return kit.Redirect(http.StatusSeeOther, "/")
	}

	idStr := chi.URLParam(kit.Request, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return errors.New("invalid product ID")
	}

	var product models.Product
	if err := db.Get().First(&product, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return kit.Render(viewerrors.Error404())
		}
		return err
	}

	allCategories := helpers.GetCategoryTree()

	// Preload existing categories for the product
	db.Get().Model(&product).Association("Categories").Find(&product.Categories)

	modal := products.EditModal(product, allCategories)

	if kit.Request.Header.Get("HX-Request") == "true" {
		return kit.Render(modal)
	}

	activePath := "/admin/products"
	sidebar := config.GetAdminSidebar()
	return RenderWithLayout(kit, layouts.AdminPage(sidebar, activePath, modal))
}

// HandleProductUpdate handles the update of a product by an admin user.
// It verifies the user's role, parses the product ID from the URL, retrieves the product from the database,
// updates the product's name and price from form values, and optionally updates the product's image if provided.
// The updated product is saved to the database. Non-admin users are redirected to the home page.
// On success, redirects to the plants listing page.
func HandleProductUpdate(kit *kit.Kit) error {
	user, ok := kit.Auth().(models.AuthUser)
	if !ok || user.Role != "admin" {
		return kit.Redirect(http.StatusSeeOther, "/")
	}

	idStr := chi.URLParam(kit.Request, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return errors.New("invalid product ID")
	}

	var product models.Product
	if err := db.Get().First(&product, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return kit.Render(viewerrors.Error404())
		}
		return err
	}

	// Parse form
	kit.Request.ParseMultipartForm(10 << 20)

	product.Name = kit.Request.FormValue("name")
	product.Description = kit.Request.FormValue("description")
	if price, err := strconv.ParseFloat(kit.Request.FormValue("price"), 64); err == nil {
		product.Price = price
	}
	promPriceStr := kit.Request.FormValue("promotion_price")
	if promPriceStr == "" {
		product.PromotionPrice = 0
	} else if promPrice, err := strconv.ParseFloat(promPriceStr, 64); err == nil {
		product.PromotionPrice = promPrice
	}
	if stock, err := strconv.Atoi(kit.Request.FormValue("stock")); err == nil {
		product.Stock = stock
	}

	// Handle optional image update
	file, header, err := kit.Request.FormFile("image")
	if err == nil {
		defer file.Close()

		uploadPath := "public/images/products"
		if err := os.MkdirAll(uploadPath, 0755); err == nil {
			ext := filepath.Ext(header.Filename)
			newFileName := fmt.Sprintf("plant_%d%s", time.Now().UnixNano(), ext)
			fullPath := filepath.Join(uploadPath, newFileName)

			dst, err := os.Create(fullPath)
			if err == nil {
				defer dst.Close()
				io.Copy(dst, file)
				// remove old image if it exists
				if len(product.Image) > 1 && product.Image[0] == '/' {
					if _, err := os.Stat(product.Image[1:]); err == nil {
						os.Remove(strings.TrimPrefix(product.Image, "/"))
					}
				}
				product.Image = "/" + fullPath
			}
		}
	}

	// Update categories
	categoryIDs := kit.Request.Form["categories"]
	if len(categoryIDs) > 0 {
		var categoriesToAssign []models.Category
		db.Get().Find(&categoriesToAssign, categoryIDs)
		db.Get().Model(&product).Association("Categories").Replace(categoriesToAssign)
	}

	db.Get().Save(&product)

	return kit.Redirect(http.StatusSeeOther, "/admin/products")
}

// HandleProductQuickView renders the quick view modal for a product.
func HandleProductQuickView(kit *kit.Kit) error {
	idStr := chi.URLParam(kit.Request, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return errors.New("invalid product ID")
	}

	var product models.Product
	if err := db.Get().Preload("Categories").First(&product, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return kit.Render(viewerrors.Error404())
		}
		return err
	}

	return kit.Render(products.QuickViewModal(product))
}

// HandleProductShow renders the standalone product detail page.
func HandleProductShow(kit *kit.Kit) error {
	idStr := chi.URLParam(kit.Request, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return errors.New("invalid product ID")
	}

	var product models.Product
	if err := db.Get().Preload("Categories").First(&product, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return kit.Render(viewerrors.Error404())
		}
		return err
	}

	cfg := config.Get()
	return RenderWithLayout(kit, products.Show(product, cfg))
}
