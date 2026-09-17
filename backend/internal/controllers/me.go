package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/coderHArsitv2/higgsfield-clone/backend/internal/middleware"
)

type Me struct{ Deps }

func NewMe(d Deps) *Me { return &Me{Deps: d} }

// Get returns the caller's profile. The frontend calls this right after login
// to turn an Auth0 identity into an app user with a credit balance.
func (h *Me) Get(c *gin.Context) {
	user := middleware.CurrentUser(c)
	keys, err := h.Keys.ProviderSet(c.Request.Context(), user.ID)
	if err != nil {
		respond(c, err)
		return
	}
	providers := make([]string, 0, len(keys))
	for id := range keys {
		providers = append(providers, id)
	}
	c.JSON(http.StatusOK, gin.H{
		"user":           user,
		"byok_providers": providers,
	})
}
