package handlers

import (
	"shopTemplate/app/types"
	layouts "shopTemplate/app/views"

	"github.com/anthdm/superkit/kit"
)

func HandlePlantsIndex(kit *kit.Kit) error {
	// In a real app, you'd fetch plants from a database here.

	return RenderWithLayout(kit, layouts.PlantsLayouts(types.AuthUser{}, 1))
	//return kit.Render(layouts.PlantsLayouts(types.AuthUser{}, 1))
}
