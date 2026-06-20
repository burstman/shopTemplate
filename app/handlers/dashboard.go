package handlers

import (
	"net/http"
	"time"
	"shopTemplate/app/config"
	"shopTemplate/app/db"
	"shopTemplate/app/models"
	"shopTemplate/app/views/dashboard"

	"github.com/anthdm/superkit/kit"
)

func HandleAdminDashboard(kit *kit.Kit) error {
	user, ok := kit.Auth().(models.AuthUser)
	if !ok || user.Role != "admin" {
		return kit.Redirect(http.StatusSeeOther, "/")
	}

	type revenueResult struct {
		Total float64
	}
	var totalRev revenueResult
	db.Get().Model(&models.Order{}).Select("COALESCE(SUM(total), 0) as total").Where("status != ?", "cancelled").Scan(&totalRev)

	// Fetch affiliate expiry
	var expiresAt *time.Time
	var affiliate models.Affiliate
	if err := db.Get().Where("affiliate_id = ?", "AFF-001").First(&affiliate).Error; err == nil {
		expiresAt = affiliate.ExpiresAt
	}

	data := dashboard.DashboardData{
		AffiliateID:  "AFF-001",
		TotalRevenue: totalRev.Total,
		ExpiresAt:    expiresAt,
	}

	cfg := config.FromContext(kit.Request.Context())
	activePath := "/admin/dashboard"
	sidebar := config.GetAdminSidebarGroups()
	content := dashboard.Index(data, cfg)
	return RenderAdminWithLayout(kit, sidebar, activePath, content)
}
