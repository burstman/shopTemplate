package handlers

import (
	"shopTemplate/app/types"
	"shopTemplate/plugins/auth"

	"github.com/anthdm/superkit/kit"
)

func HandleAuthentication(kit *kit.Kit) (kit.Auth, error) {
	resp, err := auth.AuthenticateUser(kit)
	if err != nil {
		return nil, err
	}
	userAuth := resp.(auth.Auth)

	return types.AuthUser{
		ID:       userAuth.UserID,
		Email:    userAuth.Email,
		LoggedIn: userAuth.LoggedIn,
		Role:     userAuth.Role,
	}, nil
}
