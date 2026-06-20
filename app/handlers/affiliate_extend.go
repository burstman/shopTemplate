package handlers

import (
	"net/http"
	"strconv"
	"time"
	"shopTemplate/app/db"
	"shopTemplate/app/models"

	"github.com/anthdm/superkit/kit"
)

func HandleAdminExtend(kit *kit.Kit) error {
	user, ok := kit.Auth().(models.AuthUser)
	if !ok || user.Role != "admin" {
		return kit.Redirect(http.StatusSeeOther, "/")
	}

	affiliateID := kit.Request.PathValue("affiliateID")
	if affiliateID == "" {
		kit.Response.WriteHeader(http.StatusBadRequest)
		kit.Response.Write([]byte("missing affiliate id"))
		return nil
	}

	daysStr := kit.Request.FormValue("days")
	if daysStr == "" {
		kit.Response.WriteHeader(http.StatusBadRequest)
		kit.Response.Write([]byte("missing days"))
		return nil
	}

	days, err := strconv.Atoi(daysStr)
	if err != nil || days <= 0 {
		kit.Response.WriteHeader(http.StatusBadRequest)
		kit.Response.Write([]byte("days must be a positive integer"))
		return nil
	}

	var affiliate models.Affiliate
	if err := db.Get().Where("affiliate_id = ?", affiliateID).First(&affiliate).Error; err != nil {
		kit.Response.WriteHeader(http.StatusNotFound)
		kit.Response.Write([]byte("affiliate not found"))
		return nil
	}

	now := time.Now()
	var newExpiry time.Time
	if affiliate.ExpiresAt != nil && affiliate.ExpiresAt.After(now) {
		newExpiry = affiliate.ExpiresAt.AddDate(0, 0, days)
	} else {
		newExpiry = now.AddDate(0, 0, days)
	}

	if err := db.Get().Model(&models.Affiliate{}).Where("id = ?", affiliate.ID).Update("expires_at", newExpiry).Error; err != nil {
		kit.Response.WriteHeader(http.StatusInternalServerError)
		kit.Response.Write([]byte("failed to extend subscription"))
		return nil
	}

	kit.Response.Header().Set("Content-Type", "text/html")
	kit.Response.Write([]byte(`<div class="p-3 bg-green-100 text-green-800 rounded-lg text-sm font-medium border border-green-200">Subscription extended by ` + strconv.Itoa(days) + ` days — expires ` + newExpiry.Format("Jan 02, 2006") + `</div>`))
	return nil
}
