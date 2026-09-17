package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Base struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (b *Base) BeforeCreate(*gorm.DB) error {
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}
	return nil
}

// User is provisioned on first authenticated request. Auth0 owns identity, so
// there is no password column here by design.
type User struct {
	Base
	Auth0Subject string     `gorm:"uniqueIndex;not null" json:"-"`
	Email        string     `gorm:"index" json:"email"`
	Name         string     `json:"name"`
	Picture      string     `json:"picture"`
	Credits      int        `gorm:"not null;default:100" json:"credits"`
	LastSeenAt   *time.Time `json:"last_seen_at,omitempty"`

	APIKeys     []UserAPIKey `gorm:"constraint:OnDelete:CASCADE" json:"-"`
	Generations []Generation `gorm:"constraint:OnDelete:CASCADE" json:"-"`
}

// UserAPIKey is a BYOK credential. The plaintext key never leaves the server:
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

func All() []any {
	return []any{&User{}, &UserAPIKey{}, &Generation{}, &Asset{}}
}
