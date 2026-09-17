package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/coderHArsitv2/higgsfield-clone/backend/internal/middleware"
	"github.com/coderHArsitv2/higgsfield-clone/backend/pkg/apierr"
)

type Keys struct{ Deps }

func NewKeys(d Deps) *Keys { return &Keys{Deps: d} }

// List returns masked previews only. There is deliberately no endpoint that
// returns a stored key in plaintext.
func (h *Keys) List(c *gin.Context) {
	rows, err := h.Keys.List(c.Request.Context(), middleware.CurrentUser(c).ID)
	if err != nil {
		respond(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"keys": rows})
}

type saveKeyInput struct {
	ProviderID string `json:"provider_id" binding:"required"`
	APIKey     string `json:"api_key" binding:"required"`
}

func (h *Keys) Save(c *gin.Context) {
	var in saveKeyInput
	if err := c.ShouldBindJSON(&in); err != nil {
		respond(c, apierr.BadRequest("provider_id and api_key are required"))
		return
	}
	row, err := h.Keys.Save(c.Request.Context(), middleware.CurrentUser(c).ID, in.ProviderID, in.APIKey)
	if err != nil {
		respond(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"key": row})
}

func (h *Keys) Delete(c *gin.Context) {
	if err := h.Keys.Delete(c.Request.Context(), middleware.CurrentUser(c).ID, c.Param("provider")); err != nil {
		respond(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
