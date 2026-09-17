package models

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// UserAPIKey is a BYOK credential. The plaintext never leaves the server:
// Encrypted holds AES-GCM ciphertext and Preview is a masked display value.
type UserAPIKey struct {
	Base
	UserID     uuid.UUID  `gorm:"type:uuid;index;not null;uniqueIndex:idx_user_provider" json:"user_id"`
	ProviderID string     `gorm:"not null;uniqueIndex:idx_user_provider" json:"provider_id"`
	Encrypted  string     `gorm:"not null" json:"-"`
	Preview    string     `json:"preview"`
	LastUsedAt *time.Time `json:"last_used_at,omitempty"`
}

func (UserAPIKey) TableName() string { return "user_api_keys" }

type apiKeyStore struct{ db *gorm.DB }

func (s *apiKeyStore) List(ctx context.Context, userID uuid.UUID) ([]UserAPIKey, error) {
	var out []UserAPIKey
	err := s.db.WithContext(ctx).Where("user_id = ?", userID).
		Order("provider_id").Find(&out).Error
	return out, err
}

// ByProvider returns nil, nil when the user has no key for that provider --
// a missing key is an ordinary state, not an error.
func (s *apiKeyStore) ByProvider(ctx context.Context, userID uuid.UUID, providerID string) (*UserAPIKey, error) {
	var row UserAPIKey
	err := s.db.WithContext(ctx).
		Where("user_id = ? AND provider_id = ?", userID, providerID).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

// Upsert replaces an existing key rather than erroring, which is what a user
// rotating a credential expects.
func (s *apiKeyStore) Upsert(ctx context.Context, key *UserAPIKey) error {
	return s.db.WithContext(ctx).Clauses(
		clause.OnConflict{
			Columns:   []clause.Column{{Name: "user_id"}, {Name: "provider_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"encrypted", "preview", "updated_at"}),
		},
		clause.Returning{},
	).Create(key).Error
}

func (s *apiKeyStore) Delete(ctx context.Context, userID uuid.UUID, providerID string) (bool, error) {
	res := s.db.WithContext(ctx).
		Where("user_id = ? AND provider_id = ?", userID, providerID).
		Delete(&UserAPIKey{})
	return res.RowsAffected > 0, res.Error
}

func (s *apiKeyStore) MarkUsed(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	return s.db.WithContext(ctx).Model(&UserAPIKey{}).
		Where("id = ?", id).UpdateColumn("last_used_at", &now).Error
}
