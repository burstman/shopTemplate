package handlers

import (
	"shopTemplate/app/db"
	"shopTemplate/app/models"
	layouts "shopTemplate/app/views"

	"github.com/anthdm/superkit/kit"
)

func HandlePlantsIndex(kit *kit.Kit) error {
	user, ok := kit.Auth().(models.AuthUser)
	if !ok {
		user = models.AuthUser{}
	}

	var products []models.Product

	result := db.Get().Find(&products)
	if result.Error != nil {
		return result.Error
	}

	return RenderWithLayout(kit, layouts.PlantsLayouts(user, products))
}
