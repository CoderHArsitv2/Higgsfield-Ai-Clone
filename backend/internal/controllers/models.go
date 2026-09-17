package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/coderHArsitv2/higgsfield-clone/backend/internal/middleware"
	"github.com/coderHArsitv2/higgsfield-clone/backend/internal/provider"
)

type Models struct{ Deps }

func NewModels(d Deps) *Models { return &Models{Deps: d} }

// List returns the whole catalogue, annotated with whether this caller can run
// each model. Locked models are returned rather than filtered out so the UI can
// show what exists and what it would take to unlock it.
func (h *Models) List(c *gin.Context) {
	userProviders := map[string]bool{}
	if user := middleware.CurrentUser(c); user != nil {
		set, err := h.Keys.ProviderSet(c.Request.Context(), user.ID)
		if err != nil {
			respond(c, err)
			return
		}
		userProviders = set
	}

	catalog := h.Registry.Catalog(userProviders)
	counts := map[string]int{"total": len(catalog)}
	for _, m := range catalog {
		if m.Enabled {
			counts["enabled"]++
		}
		counts[string(m.Modality)]++
	}

	c.JSON(http.StatusOK, gin.H{
		"models":    catalog,
		"providers": h.Registry.ProviderViews(userProviders),
		"counts":    counts,
	})
}

// Public is the unauthenticated catalogue used by the landing page's model
// wall. It never reflects a user's own keys.
func (h *Models) Public(c *gin.Context) {
	catalog := h.Registry.Catalog(nil)
	out := make([]gin.H, 0, len(catalog))
	for _, m := range catalog {
		out = append(out, gin.H{
			"id": m.ID, "name": m.Name, "modality": m.Modality,
			"provider_name": m.ProviderName, "description": m.Description,
			"enabled": m.Enabled, "featured": m.Featured, "tags": m.Tags,
		})
	}
	byModality := map[provider.Modality]int{}
	for _, m := range catalog {
		byModality[m.Modality]++
	}
	c.JSON(http.StatusOK, gin.H{"models": out, "by_modality": byModality})
}
