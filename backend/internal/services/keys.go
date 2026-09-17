package services

import (
	"context"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"github.com/coderHArsitv2/higgsfield-clone/backend/internal/models"
	"github.com/coderHArsitv2/higgsfield-clone/backend/internal/provider"
	"github.com/coderHArsitv2/higgsfield-clone/backend/pkg/apierr"
	"github.com/coderHArsitv2/higgsfield-clone/backend/pkg/cryptox"
)

// Keys implements bring-your-own-key. A stored key is encrypted at rest and is
// never sent back to the client -- only a masked preview is -- so a compromised
// session cannot exfiltrate the user's billing credential.
type Keys struct {
	store    models.APIKeyStore
	cipher   *cryptox.Cipher
	registry *provider.Registry
}

func NewKeys(store models.APIKeyStore, cipher *cryptox.Cipher, reg *provider.Registry) *Keys {
	return &Keys{store: store, cipher: cipher, registry: reg}
}

func (k *Keys) List(ctx context.Context, userID uuid.UUID) ([]models.UserAPIKey, error) {
	return k.store.List(ctx, userID)
}

// ProviderSet is the set of providers this user has a key for, used to decide
// which models show as unlocked.
func (k *Keys) ProviderSet(ctx context.Context, userID uuid.UUID) (map[string]bool, error) {
	rows, err := k.store.List(ctx, userID)
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
	if err := k.store.Upsert(ctx, &row); err != nil {
		return nil, err
	}
	return &row, nil
}

func (k *Keys) Delete(ctx context.Context, userID uuid.UUID, providerID string) error {
	removed, err := k.store.Delete(ctx, userID, providerID)
	if err != nil {
		return err
	}
	if !removed {
		return apierr.ErrNotFound
	}
	return nil
}

// Plaintext decrypts a user's key for a single outbound call. An empty string
// with no error means the user simply has no key for that provider.
func (k *Keys) Plaintext(ctx context.Context, userID uuid.UUID, providerID string) (string, error) {
	row, err := k.store.ByProvider(ctx, userID, providerID)
	if err != nil {
		return "", err
	}
	if row == nil {
		return "", nil
	}

	plain, err := k.cipher.Decrypt(row.Encrypted)
	if err != nil {
		// Almost always means ENCRYPTION_KEY changed under existing rows.
		return "", apierr.New(http.StatusConflict, "key_undecryptable",
			"your stored "+providerID+" key could not be decrypted; please re-add it")
	}
	_ = k.store.MarkUsed(ctx, row.ID)
	return plain, nil
}
