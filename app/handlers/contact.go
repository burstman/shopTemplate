package handlers

import (
	"shopTemplate/app/views/contact"

	"github.com/anthdm/superkit/kit"
)

func HandleContactIndex(kit *kit.Kit) error {
	return RenderWithLayout(kit, contact.Index())
}
