package handlers

import (
	"shopTemplate/app/config"
	"shopTemplate/app/helpers"
	"shopTemplate/app/models"
	"shopTemplate/app/views/layouts"

	"github.com/a-h/templ"
	"github.com/anthdm/superkit/kit"
)

func RenderWithLayout(kit *kit.Kit, content templ.Component) error {
	var user models.AuthUser
	if authedUser, ok := kit.Auth().(models.AuthUser); ok {
		user = authedUser
	}

	categories := helpers.GetCategoryTree()
	cart := helpers.GetCart(kit)

	return kit.Render(layouts.App(user, config.Get(), categories, cart.Total, content))
}
