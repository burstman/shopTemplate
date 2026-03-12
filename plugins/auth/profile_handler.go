package auth

import (
	"fmt"
	"shopTemplate/app/db"
	"shopTemplate/app/types"
	"shopTemplate/app/views/layouts"

	"github.com/anthdm/superkit/kit"
	v "github.com/anthdm/superkit/validate"
)

var profileSchema = v.Schema{
	"firstName": v.Rules(v.Min(3), v.Max(50)),
	"lastName":  v.Rules(v.Min(3), v.Max(50)),
}

type ProfileFormValues struct {
	ID        uint   `form:"id"`
	FirstName string `form:"firstName"`
	LastName  string `form:"lastName"`
	Email     string
	Success   string
}

func HandleProfileShow(kit *kit.Kit) error {
	pluginAuth := kit.Auth().(Auth)

	var user User
	if err := db.Get().First(&user, pluginAuth.UserID).Error; err != nil {
		return err
	}

	formValues := ProfileFormValues{
		ID:        user.ID,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Email:     user.Email,
	}

	appUser := types.AuthUser{
		ID:       pluginAuth.UserID,
		Email:    pluginAuth.Email,
		LoggedIn: pluginAuth.LoggedIn,
		Role:     pluginAuth.Role,
	}
	return kit.Render(layouts.App(appUser, ProfileShow(formValues)))
}

func HandleProfileUpdate(kit *kit.Kit) error {
	var values ProfileFormValues
	errors, ok := v.Request(kit.Request, &values, profileSchema)
	if !ok {
		return kit.Render(ProfileForm(values, errors))
	}

	pluginAuth := kit.Auth().(Auth)
	if pluginAuth.UserID != values.ID {
		return fmt.Errorf("unauthorized request for profile %d", values.ID)
	}
	err := db.Get().Model(&User{}).
		Where("id = ?", pluginAuth.UserID).
		Updates(&User{
			FirstName: values.FirstName,
			LastName:  values.LastName,
		}).Error
	if err != nil {
		return err
	}

	values.Success = "Profile successfully updated!"
	values.Email = pluginAuth.Email

	return kit.Render(ProfileForm(values, v.Errors{}))
}
