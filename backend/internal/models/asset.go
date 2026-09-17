package models

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Asset is one output file of a generation.
type Asset struct {
	Base
	GenerationID uuid.UUID `gorm:"type:uuid;index;not null" json:"generation_id"`
	Kind         string    `gorm:"not null" json:"kind"`
	URL          string    `gorm:"type:text;not null" json:"url"`
	ThumbnailURL string    `gorm:"type:text" json:"thumbnail_url,omitempty"`
	Width        int       `json:"width,omitempty"`
	Height       int       `json:"height,omitempty"`
	DurationMS   int       `json:"duration_ms,omitempty"`
}

type assetStore struct{ db *gorm.DB }

func (s *assetStore) CreateMany(ctx context.Context, assets []Asset) error {
	if len(assets) == 0 {
		return nil
	}
	return s.db.WithContext(ctx).Create(&assets).Error
}
