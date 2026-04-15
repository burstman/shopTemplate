package handlers

import (
	"shopTemplate/plugins/auth"

	"github.com/anthdm/superkit/kit"
)

func HandleAuthentication(kit *kit.Kit) (kit.Auth, error) {
	return auth.AuthenticateUser(kit)
}
