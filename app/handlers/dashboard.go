package handlers

import (
	"math"
	"net/http"
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

	var totalOrders int64
	db.Get().Model(&models.Order{}).Count(&totalOrders)

	type revenueResult struct {
		Total float64
	}
	var totalRev revenueResult
	db.Get().Model(&models.Order{}).Select("COALESCE(SUM(total), 0) as total").Where("status != ?", "cancelled").Scan(&totalRev)

	type commissionResult struct {
		Total   float64
		Pending float64
		Paid    float64
	}
	var comm commissionResult
	db.Get().Model(&models.Order{}).Select(`
		COALESCE(SUM(platform_commission), 0) as total,
		COALESCE(SUM(platform_commission) FILTER (WHERE commission_status = 'pending'), 0) as pending,
		COALESCE(SUM(platform_commission) FILTER (WHERE commission_status = 'paid'), 0) as paid
	`).Where("is_test = ?", false).Scan(&comm)

	type statusCount struct {
		Status string
		Count  int64
	}
	var statusCounts []statusCount
	db.Get().Model(&models.Order{}).Select("status, COUNT(*) as count").Group("status").Scan(&statusCounts)

	ordersByStatus := make(map[string]int64)
	for _, sc := range statusCounts {
		ordersByStatus[sc.Status] = sc.Count
	}

	var recentOrders []models.Order
	db.Get().Preload("Items").Order("created_at desc").Limit(10).Find(&recentOrders)

	totalCommission := math.Round(comm.Total*100) / 100
	pendingCommission := math.Round(comm.Pending*100) / 100
	paidCommission := math.Round(comm.Paid*100) / 100
	totalRevenue := math.Round(totalRev.Total*100) / 100

	data := dashboard.DashboardData{
		TotalOrders:       totalOrders,
		TotalRevenue:      totalRevenue,
		TotalCommission:   totalCommission,
		PendingCommission: pendingCommission,
		PaidCommission:    paidCommission,
		OrdersByStatus:    ordersByStatus,
		RecentOrders:      recentOrders,
	}

	cfg := config.Get()
	activePath := "/admin/dashboard"
	sidebar := config.GetAdminSidebar()
	content := dashboard.Index(data, cfg)
	return RenderAdminWithLayout(kit, sidebar, activePath, content)
}
