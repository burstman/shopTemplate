package handlers

import (
	"net/http"
	"shopTemplate/app/db"
	"shopTemplate/app/models"
	"shopTemplate/app/types"
	"shopTemplate/app/views/components"

	"github.com/anthdm/superkit/kit"
)

func HandleConfigurationIndex(kit *kit.Kit) error {
	user, ok := kit.Auth().(types.AuthUser)
	if !ok || user.Role != "admin" {
		return kit.Redirect(http.StatusSeeOther, "/")
	}

	// Fetch settings from DB
	var settings []models.Setting
	db.Get().Find(&settings)

	// Convert to map for easy access in view
	configMap := make(map[string]string)
	for _, s := range settings {
		configMap[s.Key] = s.Value
	}

	return RenderWithLayout(kit, components.Configuration(configMap))
}

func HandleConfigurationUpdate(kit *kit.Kit) error {
	user, ok := kit.Auth().(types.AuthUser)
	if !ok || user.Role != "admin" {
		return kit.Redirect(http.StatusSeeOther, "/")
	}

	// Parse form values
	err := kit.Request.ParseForm()
	if err != nil {
		return err
	}

	// Iterate over posted values and update/create settings
	for key, values := range kit.Request.PostForm {
		if len(values) > 0 {
			// Upsert logic: Save key-value pair
			db.Get().Where(models.Setting{Key: key}).Assign(models.Setting{Value: values[0]}).FirstOrCreate(&models.Setting{})
		}
	}

	return kit.Redirect(http.StatusSeeOther, "/configuration")
}
