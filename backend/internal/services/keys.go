package services

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/coderHArsitv2/higgsfield-clone/backend/internal/models"
	"github.com/coderHArsitv2/higgsfield-clone/backend/internal/provider"
	"github.com/coderHArsitv2/higgsfield-clone/backend/pkg/apierr"
	"github.com/coderHArsitv2/higgsfield-clone/backend/pkg/cryptox"
)

// Keys implements bring-your-own-key. A stored key is encrypted at rest and is
// never sent back to the client -- only a masked preview is -- so a compromised
// session cannot exfiltrate the user's billing credential.
type Keys struct {
	db       *gorm.DB
	cipher   *cryptox.Cipher
	registry *provider.Registry
}

func NewKeys(db *gorm.DB, cipher *cryptox.Cipher, reg *provider.Registry) *Keys {
	return &Keys{db: db, cipher: cipher, registry: reg}
}

func (k *Keys) List(ctx context.Context, userID uuid.UUID) ([]models.UserAPIKey, error) {
	var out []models.UserAPIKey
	err := k.db.WithContext(ctx).Where("user_id = ?", userID).Order("provider_id").Find(&out).Error
	return out, err
}

// ProviderSet is the set of providers this user has a key for, used to decide
// which models show as unlocked.
func (k *Keys) ProviderSet(ctx context.Context, userID uuid.UUID) (map[string]bool, error) {
	rows, err := k.List(ctx, userID)
	if err != nil {
		return nil, err
	}
	set := make(map[string]bool, len(rows))
	for _, r := range rows {
		set[r.ProviderID] = true
	}
	return set, nil
}

func (k *Keys) Save(ctx context.Context, userID uuid.UUID, providerID, raw string) (*models.UserAPIKey, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, apierr.BadRequest("api key cannot be empty")
	}
	p, ok := k.registry.Provider(providerID)
	if !ok {
		return nil, apierr.BadRequest("unknown provider " + providerID)
	}
	if p.EnvKey() == "" {
		return nil, apierr.BadRequest(p.Name() + " does not take an API key")
	}

	enc, err := k.cipher.Encrypt(raw)
	if err != nil {
		return nil, err
	}
	row := models.UserAPIKey{
		UserID:     userID,
		ProviderID: providerID,
		Encrypted:  enc,
		Preview:    cryptox.Mask(raw),
	}
	// Re-saving replaces the key rather than erroring, which is what a user
	// rotating a credential expects.
	err = k.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}, {Name: "provider_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"encrypted", "preview", "updated_at"}),
	}).Create(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (k *Keys) Delete(ctx context.Context, userID uuid.UUID, providerID string) error {
	res := k.db.WithContext(ctx).
		Where("user_id = ? AND provider_id = ?", userID, providerID).
		Delete(&models.UserAPIKey{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return apierr.ErrNotFound
	}
	return nil
}

// Plaintext decrypts a user's key for a single outbound call.
func (k *Keys) Plaintext(ctx context.Context, userID uuid.UUID, providerID string) (string, error) {
	var row models.UserAPIKey
	err := k.db.WithContext(ctx).
		Where("user_id = ? AND provider_id = ?", userID, providerID).
		First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	plain, err := k.cipher.Decrypt(row.Encrypted)
	if err != nil {
		// Almost always means ENCRYPTION_KEY changed under existing rows.
		return "", apierr.New(http.StatusConflict, "key_undecryptable",
			"your stored "+providerID+" key could not be decrypted; please re-add it")
	}
	now := time.Now()
	k.db.WithContext(ctx).Model(&row).UpdateColumn("last_used_at", &now)
	return plain, nil
}
