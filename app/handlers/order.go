package handlers

import (
	"errors"
	"fmt"
	"math"
	"net/http"
	"shopTemplate/app/config"
	"shopTemplate/app/db"
	"shopTemplate/app/models"
	viewerrors "shopTemplate/app/views/errors"
	"shopTemplate/app/views/layouts"
	"shopTemplate/app/views/orders"
	"strconv"

	"github.com/anthdm/superkit/kit"
	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
)

func HandleAdminOrdersIndex(kit *kit.Kit) error {
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
	db.Get().Model(&models.Order{}).Count(&total)
	totalPages := int(math.Ceil(float64(total) / float64(perPage)))
	offset := (page - 1) * perPage

	var ordersList []models.Order
	db.Get().Order("created_at desc").Limit(perPage).Offset(offset).Find(&ordersList)

	activePath := "/admin/orders"
	sidebar := config.GetAdminSidebar()
	content := orders.Index(ordersList, page, totalPages)
	return RenderWithLayout(kit, layouts.AdminPage(sidebar, activePath, content))
}

func HandleAdminOrderShow(kit *kit.Kit) error {
	user, ok := kit.Auth().(models.AuthUser)
	if !ok || user.Role != "admin" {
		return kit.Redirect(http.StatusSeeOther, "/")
	}

	idStr := chi.URLParam(kit.Request, "id")
	id, _ := strconv.Atoi(idStr)

	var order models.Order
	if err := db.Get().Preload("Items.Product").First(&order, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return kit.Render(viewerrors.Error404())
		}
		return err
	}

	activePath := "/admin/orders"
	sidebar := config.GetAdminSidebar()
	content := orders.Show(order)
	return RenderWithLayout(kit, layouts.AdminPage(sidebar, activePath, content))
}

func HandleAdminOrderUpdateStatus(kit *kit.Kit) error {
	user, ok := kit.Auth().(models.AuthUser)
	if !ok || user.Role != "admin" {
		return kit.Redirect(http.StatusSeeOther, "/")
	}

	idStr := chi.URLParam(kit.Request, "id")
	id, _ := strconv.Atoi(idStr)
	status := kit.Request.FormValue("status")

	if err := db.Get().Model(&models.Order{}).Where("id = ?", id).Update("status", status).Error; err != nil {
		return err
	}

	return kit.Redirect(http.StatusSeeOther, fmt.Sprintf("/admin/orders/%d", id))
}
