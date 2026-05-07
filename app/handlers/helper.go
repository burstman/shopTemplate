package handlers

import (
	"shopTemplate/app/config"
	"shopTemplate/app/helpers"
	"shopTemplate/app/models"
	"shopTemplate/app/views/layouts"

	"github.com/a-h/templ"
	"github.com/anthdm/superkit/kit"
	"github.com/gorilla/csrf"
)

func RenderWithLayout(kit *kit.Kit, content templ.Component) error {
	var user models.AuthUser
	if authedUser, ok := kit.Auth().(models.AuthUser); ok {
		user = authedUser
	}

	categories := helpers.GetCategoryTree()
	cart := helpers.GetCart(kit)
	csrfToken := csrf.Token(kit.Request)

	return kit.Render(layouts.App(user, config.Get(), categories, cart.Total, content, csrfToken))
}

func RenderAdminWithLayout(kit *kit.Kit, sidebar []config.MenuItem, activePath string, content templ.Component) error {
	var user models.AuthUser
	if authedUser, ok := kit.Auth().(models.AuthUser); ok {
		user = authedUser
	}

	cfg := config.Get()
	categories := helpers.GetCategoryTree()
	cart := helpers.GetCart(kit)
	csrfToken := csrf.Token(kit.Request)

	return kit.Render(layouts.AdminPage(user, cfg, categories, cart.Total, sidebar, activePath, content, csrfToken))
}
