package handlers

import (
	"net/http"

	"shopTemplate/app/helpers"
	"shopTemplate/app/models"
	"shopTemplate/app/views/legal"

	"github.com/anthdm/superkit/kit"
)

func HandleHealthCheck(kit *kit.Kit) error {
	return kit.Text(http.StatusOK, "OK")
}

func HandlePrivacyPolicy(kit *kit.Kit) error {
	user, _ := kit.Auth().(models.AuthUser)
	categories := helpers.GetCategoryTree()
	cart := helpers.GetCart(kit)
	return kit.Render(legal.PrivacyPolicy(user, categories, cart.Total))
}

func HandleDataDeletion(kit *kit.Kit) error {
	user, _ := kit.Auth().(models.AuthUser)
	categories := helpers.GetCategoryTree()
	cart := helpers.GetCart(kit)
	return kit.Render(legal.DataDeletion(user, categories, cart.Total))
}
