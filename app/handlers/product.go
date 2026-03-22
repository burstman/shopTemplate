package handlers

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"shopTemplate/app/config"
	"shopTemplate/app/db"
	"shopTemplate/app/models"
	"shopTemplate/app/views/layouts"
	"shopTemplate/app/views/products"

	"github.com/anthdm/superkit/kit"
	"github.com/anthdm/superkit/validate"
	"github.com/go-chi/chi/v5"
)

func HandleProductNew(kit *kit.Kit) error {
	user, ok := kit.Auth().(models.AuthUser)
	if !ok || user.Role != "admin" {
		return kit.Redirect(http.StatusSeeOther, "/")
	}

	categories := getCategoryTree()

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
	if err != nil {
		errors.Add("image", "Image is required")
	}
	if len(categoryIDs) == 0 {
		errors.Add("categories", "Please select at least one category")
	}

	if len(errors) > 0 {
		categories := getCategoryTree()
		activePath := "/products/new"
		sidebar := config.GetAdminSidebar()
		form := products.CreateForm{
			Values: map[string]string{
				"name":  name,
				"price": priceStr,
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
		Name:  name,
		Price: price,
		Image: "/" + fullPath,
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
	return kit.Redirect(http.StatusSeeOther, "/products")
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

	// Soft delete the product
	result := db.Get().Delete(&models.Product{}, id)
	result = db.Get().Delete(&product)
	if result.Error != nil {
		return result.Error
	}

	return kit.Text(http.StatusOK, "")
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
		return err
	}

	allCategories := getCategoryTree()

	// Preload existing categories for the product
	db.Get().Model(&product).Association("Categories").Find(&product.Categories)

	return RenderWithLayout(kit, products.EditModal(product, allCategories))
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
		return err
	}

	// Parse form
	kit.Request.ParseMultipartForm(10 << 20)

	product.Name = kit.Request.FormValue("name")
	if price, err := strconv.ParseFloat(kit.Request.FormValue("price"), 64); err == nil {
		product.Price = price
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

	return kit.Redirect(http.StatusSeeOther, "/products")
}
