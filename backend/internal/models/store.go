package models

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/coderHArsitv2/higgsfield-clone/backend/pkg/jwtx"
)

// The interfaces every caller outside this package depends on. Services take
// these rather than a *gorm.DB, so a service can be exercised against a fake
// and no SQL leaks upward.

type UserStore interface {
	// Upsert provisions the row for an Auth0 identity and returns it as the
	// database holds it.
	Upsert(ctx context.Context, claims jwtx.Claims) (*User, error)
	ByID(ctx context.Context, id uuid.UUID) (*User, error)
	// Spend deducts credits only if the balance covers it, reporting whether it
	// did. The check and the write are one statement so two concurrent
	// generations cannot both pass and overdraw the account.
	Spend(ctx context.Context, id uuid.UUID, amount int) (bool, error)
	Refund(ctx context.Context, id uuid.UUID, amount int) error
}

type APIKeyStore interface {
	List(ctx context.Context, userID uuid.UUID) ([]UserAPIKey, error)
	ByProvider(ctx context.Context, userID uuid.UUID, providerID string) (*UserAPIKey, error)
	Upsert(ctx context.Context, key *UserAPIKey) error
	Delete(ctx context.Context, userID uuid.UUID, providerID string) (bool, error)
	MarkUsed(ctx context.Context, id uuid.UUID) error
}

type GenerationFilter struct {
	Status   string
	Modality string
	Limit    int
	Offset   int
}

type GenerationStore interface {
	Create(ctx context.Context, gen *Generation) error
	ByID(ctx context.Context, userID, id uuid.UUID) (*Generation, error)
	List(ctx context.Context, userID uuid.UUID, f GenerationFilter) ([]Generation, int64, error)
	// Pending returns queued and running jobs for the worker to advance.
	Pending(ctx context.Context, limit int) ([]Generation, error)
	// ClaimForSubmit moves a job from queued to running, reporting whether this
	// caller is the one that moved it. Two workers cannot both submit a job.
	ClaimForSubmit(ctx context.Context, id uuid.UUID) (bool, error)
	Requeue(ctx context.Context, id uuid.UUID) error
	SetExternalID(ctx context.Context, id uuid.UUID, externalID string) error
	// Finish writes a terminal status, reporting whether this caller performed
	// the transition. Only that caller may refund, or a double poll refunds
	// twice.
	Finish(ctx context.Context, id uuid.UUID, status GenerationStatus, reason string) (bool, error)
	Cancel(ctx context.Context, userID, id uuid.UUID) (*Generation, error)
}

type AssetStore interface {
	CreateMany(ctx context.Context, assets []Asset) error
}

// Stores is the set handed to services. Tx runs a function with every store
// bound to one transaction, which is what credit accounting needs.
type Stores struct {
	Users       UserStore
	Keys        APIKeyStore
	Generations GenerationStore
	Assets      AssetStore

	db *gorm.DB
}

func New(db *gorm.DB) *Stores {
	return &Stores{
		Users:       &userStore{db},
		Keys:        &apiKeyStore{db},
		Generations: &generationStore{db},
		Assets:      &assetStore{db},
		db:          db,
	}
}

func (s *Stores) Tx(ctx context.Context, fn func(tx *Stores) error) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(New(tx))
	})
}

// DB is the escape hatch for health checks and migrations only.
func (s *Stores) DB() *gorm.DB { return s.db }
