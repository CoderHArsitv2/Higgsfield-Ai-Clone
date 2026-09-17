package models

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type GenerationStatus string

const (
	StatusQueued    GenerationStatus = "queued"
	StatusRunning   GenerationStatus = "running"
	StatusSucceeded GenerationStatus = "succeeded"
	StatusFailed    GenerationStatus = "failed"
	StatusCanceled  GenerationStatus = "canceled"
)

func (s GenerationStatus) Terminal() bool {
	return s == StatusSucceeded || s == StatusFailed || s == StatusCanceled
}

// Generation is one job. ExternalID is the provider's own job handle, which is
// what the polling worker uses to ask for progress.
type Generation struct {
	Base
	UserID      uuid.UUID        `gorm:"type:uuid;index;not null" json:"user_id"`
	ProviderID  string           `gorm:"index;not null" json:"provider_id"`
	ModelID     string           `gorm:"index;not null" json:"model_id"`
	Modality    string           `gorm:"index;not null" json:"modality"`
	Prompt      string           `gorm:"type:text" json:"prompt"`
	Params      JSON             `gorm:"type:jsonb" json:"params"`
	Status      GenerationStatus `gorm:"index;not null;default:'queued'" json:"status"`
	ExternalID  string           `gorm:"index" json:"-"`
	Error       string           `gorm:"type:text" json:"error,omitempty"`
	UsedOwnKey  bool             `gorm:"not null;default:false" json:"used_own_key"`
	Attempts    int              `gorm:"not null;default:0" json:"-"`
	StartedAt   *time.Time       `json:"started_at,omitempty"`
	CompletedAt *time.Time       `json:"completed_at,omitempty"`

	Assets []Asset `gorm:"constraint:OnDelete:CASCADE" json:"assets"`
}

type generationStore struct{ db *gorm.DB }

func (s *generationStore) Create(ctx context.Context, gen *Generation) error {
	return s.db.WithContext(ctx).Create(gen).Error
}

func (s *generationStore) ByID(ctx context.Context, userID, id uuid.UUID) (*Generation, error) {
	var row Generation
	err := s.db.WithContext(ctx).Preload("Assets").
		Where("id = ? AND user_id = ?", id, userID).First(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (s *generationStore) List(ctx context.Context, userID uuid.UUID, f GenerationFilter) ([]Generation, int64, error) {
	if f.Limit <= 0 || f.Limit > 100 {
		f.Limit = 24
	}
	q := s.db.WithContext(ctx).Model(&Generation{}).Where("user_id = ?", userID)
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

	var out []Generation
	err := q.Preload("Assets").Order("created_at DESC").
		Limit(f.Limit).Offset(f.Offset).Find(&out).Error
	return out, total, err
}

func (s *generationStore) Pending(ctx context.Context, limit int) ([]Generation, error) {
	var out []Generation
	err := s.db.WithContext(ctx).
		Where("status IN ?", []GenerationStatus{StatusQueued, StatusRunning}).
		Order("created_at ASC").Limit(limit).Find(&out).Error
	return out, err
}

func (s *generationStore) ClaimForSubmit(ctx context.Context, id uuid.UUID) (bool, error) {
	now := time.Now()
	res := s.db.WithContext(ctx).Model(&Generation{}).
		Where("id = ? AND status = ?", id, StatusQueued).
		Updates(map[string]any{
			"status":     StatusRunning,
			"started_at": &now,
			"attempts":   gorm.Expr("attempts + 1"),
		})
	return res.RowsAffected == 1, res.Error
}

func (s *generationStore) Requeue(ctx context.Context, id uuid.UUID) error {
	return s.db.WithContext(ctx).Model(&Generation{}).Where("id = ?", id).
		Updates(map[string]any{"status": StatusQueued, "started_at": nil}).Error
}

func (s *generationStore) SetExternalID(ctx context.Context, id uuid.UUID, externalID string) error {
	return s.db.WithContext(ctx).Model(&Generation{}).
		Where("id = ?", id).UpdateColumn("external_id", externalID).Error
}

func (s *generationStore) Finish(ctx context.Context, id uuid.UUID, status GenerationStatus, reason string) (bool, error) {
	now := time.Now()
	res := s.db.WithContext(ctx).Model(&Generation{}).
		Where("id = ? AND status NOT IN ?", id,
			[]GenerationStatus{StatusSucceeded, StatusFailed, StatusCanceled}).
		Updates(map[string]any{
			"status":       status,
			"error":        reason,
			"completed_at": &now,
		})
	return res.RowsAffected == 1, res.Error
}

func (s *generationStore) Cancel(ctx context.Context, userID, id uuid.UUID) (*Generation, error) {
	if _, err := s.ByID(ctx, userID, id); err != nil {
		return nil, err
	}
	now := time.Now()
	res := s.db.WithContext(ctx).Model(&Generation{}).
		Where("id = ? AND user_id = ? AND status IN ?", id, userID,
			[]GenerationStatus{StatusQueued, StatusRunning}).
		Updates(map[string]any{"status": StatusCanceled, "completed_at": &now})
	if res.Error != nil {
		return nil, res.Error
	}
	if res.RowsAffected == 0 {
		return nil, ErrAlreadyFinished
	}
	return s.ByID(ctx, userID, id)
}
