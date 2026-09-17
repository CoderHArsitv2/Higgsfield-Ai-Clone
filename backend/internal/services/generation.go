package services

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/coderHArsitv2/higgsfield-clone/backend/internal/models"
	"github.com/coderHArsitv2/higgsfield-clone/backend/internal/provider"
	"github.com/coderHArsitv2/higgsfield-clone/backend/pkg/apierr"
)

type Generations struct {
	db       *gorm.DB
	registry *provider.Registry
	keys     *Keys
}

func NewGenerations(db *gorm.DB, reg *provider.Registry, keys *Keys) *Generations {
	return &Generations{db: db, registry: reg, keys: keys}
}

type CreateInput struct {
	ModelID   string         `json:"model_id" binding:"required"`
	Prompt    string         `json:"prompt" binding:"required"`
	Params    map[string]any `json:"params"`
	RefImages []string       `json:"ref_images"`
}

func (g *Generations) Create(ctx context.Context, user *models.User, in CreateInput) (*models.Generation, error) {
	prompt := strings.TrimSpace(in.Prompt)
	if prompt == "" {
		return nil, apierr.BadRequest("prompt cannot be empty")
	}
	if len(prompt) > 5000 {
		return nil, apierr.BadRequest("prompt is too long (max 5000 characters)")
	}

	p, spec, ok := g.registry.Model(in.ModelID)
	if !ok {
		return nil, apierr.BadRequest("unknown model " + in.ModelID)
	}
	if len(in.RefImages) > spec.RefImages {
		return nil, apierr.BadRequest(fmt.Sprintf("%s accepts at most %d reference images", spec.Name, spec.RefImages))
	}

	userKey, err := g.keys.Plaintext(ctx, user.ID, p.ID())
	if err != nil {
		return nil, err
	}
	_, ownKey, err := g.registry.ResolveKey(p.ID(), userKey)
	if err != nil {
		return nil, apierr.Unprocessable(spec.Name + " is locked. Add a " + p.Name() + " API key in settings to use it.")
	}

	// A user paying with their own key does not spend platform credits.
	if !ownKey && user.Credits < spec.CreditCost {
		return nil, apierr.Unprocessable(fmt.Sprintf(
			"not enough credits: %s costs %d, you have %d. Add your own %s key to bypass credits.",
			spec.Name, spec.CreditCost, user.Credits, p.Name()))
	}

	params, err := json.Marshal(normaliseParams(spec, in.Params))
	if err != nil {
		return nil, err
	}

	gen := models.Generation{
		UserID:     user.ID,
		ProviderID: p.ID(),
		ModelID:    spec.ID,
		Modality:   string(spec.Modality),
		Prompt:     prompt,
		Params:     models.JSON(params),
		Status:     models.StatusQueued,
		UsedOwnKey: ownKey,
	}

	err = g.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&gen).Error; err != nil {
			return err
		}
		if ownKey {
			return nil
		}
		// Conditional decrement: two concurrent submissions cannot both pass
		// the balance check above and overdraw the account.
		res := tx.Model(&models.User{}).
			Where("id = ? AND credits >= ?", user.ID, spec.CreditCost).
			UpdateColumn("credits", gorm.Expr("credits - ?", spec.CreditCost))
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return apierr.Unprocessable("not enough credits")
		}
		user.Credits -= spec.CreditCost
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &gen, nil
}

// normaliseParams drops anything the model did not declare, so a crafted
// request cannot smuggle arbitrary fields into an upstream provider call.
func normaliseParams(spec provider.ModelSpec, in map[string]any) map[string]any {
	out := map[string]any{}
	for _, p := range spec.Params {
		if v, ok := in[p.Key]; ok && v != nil && v != "" {
			out[p.Key] = v
			continue
		}
		if p.Default != nil {
			out[p.Key] = p.Default
		}
	}
	return out
}

type ListFilter struct {
	Status   string
	Modality string
	Limit    int
	Offset   int
}

func (g *Generations) List(ctx context.Context, userID uuid.UUID, f ListFilter) ([]models.Generation, int64, error) {
	if f.Limit <= 0 || f.Limit > 100 {
		f.Limit = 24
	}
	q := g.db.WithContext(ctx).Model(&models.Generation{}).Where("user_id = ?", userID)
	if f.Status != "" {
		q = q.Where("status = ?", f.Status)
	}
	if f.Modality != "" {
		q = q.Where("modality = ?", f.Modality)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var out []models.Generation
	err := q.Preload("Assets").Order("created_at DESC").
		Limit(f.Limit).Offset(f.Offset).Find(&out).Error
	return out, total, err
}

func (g *Generations) Get(ctx context.Context, userID, id uuid.UUID) (*models.Generation, error) {
	var gen models.Generation
	err := g.db.WithContext(ctx).Preload("Assets").
		Where("id = ? AND user_id = ?", id, userID).First(&gen).Error
	if err != nil {
		return nil, apierr.ErrNotFound
	}
	return &gen, nil
}

func (g *Generations) Cancel(ctx context.Context, userID, id uuid.UUID) (*models.Generation, error) {
	gen, err := g.Get(ctx, userID, id)
	if err != nil {
		return nil, err
	}
	if gen.Status.Terminal() {
		return nil, apierr.Conflict("this generation has already finished")
	}
	now := time.Now()
	gen.Status = models.StatusCanceled
	gen.CompletedAt = &now
	if err := g.db.WithContext(ctx).Model(gen).
		Updates(map[string]any{"status": gen.Status, "completed_at": gen.CompletedAt}).Error; err != nil {
		return nil, err
	}
	return gen, nil
}
