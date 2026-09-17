package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/coderHArsitv2/higgsfield-clone/backend/internal/middleware"
	"github.com/coderHArsitv2/higgsfield-clone/backend/internal/services"
	"github.com/coderHArsitv2/higgsfield-clone/backend/pkg/apierr"
)

type Generations struct{ Deps }

func NewGenerations(d Deps) *Generations { return &Generations{Deps: d} }

// Create accepts a job and returns immediately with a queued generation. The
// client then polls; the worker does the slow part.
func (h *Generations) Create(c *gin.Context) {
	var in services.CreateInput
	if err := c.ShouldBindJSON(&in); err != nil {
		respond(c, apierr.BadRequest("model_id and prompt are required"))
		return
	}
	user := middleware.CurrentUser(c)
	gen, err := h.Generations.Create(c.Request.Context(), user, in)
	if err != nil {
		respond(c, err)
		return
	}
	c.JSON(http.StatusAccepted, gin.H{"generation": gen, "credits_remaining": user.Credits})
}

func (h *Generations) List(c *gin.Context) {
	user := middleware.CurrentUser(c)
	limit, _ := strconv.Atoi(c.Query("limit"))
	offset, _ := strconv.Atoi(c.Query("offset"))

	rows, total, err := h.Generations.List(c.Request.Context(), user.ID, services.ListFilter{
		Status:   c.Query("status"),
		Modality: c.Query("modality"),
		Limit:    limit,
		Offset:   offset,
	})
	if err != nil {
		respond(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"generations": rows, "total": total})
}

func (h *Generations) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respond(c, apierr.BadRequest("invalid generation id"))
		return
	}
	gen, err := h.Generations.Get(c.Request.Context(), middleware.CurrentUser(c).ID, id)
	if err != nil {
		respond(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"generation": gen})
}

func (h *Generations) Cancel(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respond(c, apierr.BadRequest("invalid generation id"))
		return
	}
	gen, err := h.Generations.Cancel(c.Request.Context(), middleware.CurrentUser(c).ID, id)
	if err != nil {
		respond(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"generation": gen})
}
