package handlers

import (
	"shopTemplate/app/types"
	"shopTemplate/app/views/layouts"

	"github.com/a-h/templ"
	"github.com/anthdm/superkit/kit"
)

func RenderWithLayout(kit *kit.Kit, content templ.Component) error {
	isLoggedIn := kit.Auth().Check()

	var user types.AuthUser
	if isLoggedIn {
		// Get the authenticated user and extract the role
		if authedUser, ok := kit.Auth().(types.AuthUser); ok {
			user = authedUser
		}
	}

	return kit.Render(layouts.App(user))
}
