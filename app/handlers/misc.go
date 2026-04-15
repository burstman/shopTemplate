package handlers

import (
	"net/http"

	"github.com/anthdm/superkit/kit"
)

func HandleHealthCheck(kit *kit.Kit) error {
	return kit.Text(http.StatusOK, "OK")
}
