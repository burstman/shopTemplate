package handlers

import (
	"shopTemplate/app/views/landing"

	"github.com/anthdm/superkit/kit"
)

func HandleLandingIndex(kit *kit.Kit) error {
	return RenderWithLayout(kit, landing.Index())
}
