package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"github.com/google/uuid"

	"github.com/coderHArsitv2/higgsfield-clone/backend/internal/models"
	"github.com/coderHArsitv2/higgsfield-clone/backend/internal/provider"
	"github.com/coderHArsitv2/higgsfield-clone/backend/pkg/apierr"
)

type Generations struct {
	stores   *models.Stores
	registry *provider.Registry
	keys     *Keys
	log      *slog.Logger
}

func NewGenerations(stores *models.Stores, reg *provider.Registry, keys *Keys, log *slog.Logger) *Generations {
	return &Generations{stores: stores, registry: reg, keys: keys, log: log}
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
		return nil, apierr.BadRequest(fmt.Sprintf(
			"%s accepts at most %d reference images", spec.Name, spec.RefImages))
	}

	userKey, err := g.keys.Plaintext(ctx, user.ID, p.ID())
	if err != nil {
		return nil, err
	}
	_, ownKey, err := g.registry.ResolveKey(p.ID(), userKey)
	if err != nil {
		g.log.Warn("generation refused: no credential",
			"user", user.ID, "model", spec.ID, "provider", p.ID(),
			"has_user_key", userKey != "")
		return nil, apierr.Unprocessable(spec.Name + " is locked. Add a " +
			p.Name() + " API key in settings to use it.")
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

	err = g.stores.Tx(ctx, func(tx *models.Stores) error {
		if err := tx.Generations.Create(ctx, &gen); err != nil {
			return err
		}
		if ownKey {
			return nil
		}
		paid, err := tx.Users.Spend(ctx, user.ID, spec.CreditCost)
		if err != nil {
			return err
		}
		if !paid {
			return apierr.Unprocessable("not enough credits")
		}
		user.Credits -= spec.CreditCost
		return nil
	})
	if err != nil {
		return nil, err
	}

	g.log.Info("generation queued",
		"generation", gen.ID, "user", user.ID, "model", gen.ModelID,
		"provider", gen.ProviderID, "modality", gen.Modality,
		"own_key", ownKey, "cost", spec.CreditCost,
		"credits_left", user.Credits, "prompt_chars", len(prompt))

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

func (g *Generations) List(ctx context.Context, userID uuid.UUID, f models.GenerationFilter) ([]models.Generation, int64, error) {
	return g.stores.Generations.List(ctx, userID, f)
}

func (g *Generations) Get(ctx context.Context, userID, id uuid.UUID) (*models.Generation, error) {
	gen, err := g.stores.Generations.ByID(ctx, userID, id)
	if err != nil {
		return nil, apierr.ErrNotFound
	}
	return gen, nil
}

func (g *Generations) Cancel(ctx context.Context, userID, id uuid.UUID) (*models.Generation, error) {
	gen, err := g.stores.Generations.Cancel(ctx, userID, id)
	switch {
	case err == models.ErrAlreadyFinished:
		return nil, apierr.Conflict("this generation has already finished")
	case err != nil && models.IsNotFound(err):
		return nil, apierr.ErrNotFound
	case err != nil:
		return nil, err
	}
	g.log.Info("generation canceled", "generation", id, "user", userID)
	return gen, nil
}
