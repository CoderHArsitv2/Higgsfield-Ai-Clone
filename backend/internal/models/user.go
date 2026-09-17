package models

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/coderHArsitv2/higgsfield-clone/backend/pkg/jwtx"
)

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

type userStore struct{ db *gorm.DB }

func (s *userStore) Upsert(ctx context.Context, claims jwtx.Claims) (*User, error) {
	now := time.Now()
	row := User{
		Auth0Subject: claims.Subject,
		Email:        claims.Email,
		Name:         claims.Name,
		Picture:      claims.Picture,
		LastSeenAt:   &now,
	}

	// Profile fields are refreshed from the token; credits are deliberately not
	// in this list, so a re-login can never top someone up.
	//
	// The RETURNING clause is load-bearing. BeforeCreate mints a fresh UUID on
	// every call; on a conflict Postgres keeps the existing row's id, and
	// without RETURNING the struct comes back holding an id that was never
	// written -- which then fails every foreign key that references it.
	err := s.db.WithContext(ctx).Clauses(
		clause.OnConflict{
			Columns: []clause.Column{{Name: "auth0_subject"}},
			DoUpdates: clause.AssignmentColumns(
				[]string{"email", "name", "picture", "last_seen_at", "updated_at"}),
		},
		clause.Returning{},
	).Create(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (s *userStore) ByID(ctx context.Context, id uuid.UUID) (*User, error) {
	var row User
	if err := s.db.WithContext(ctx).First(&row, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &row, nil
}

func (s *userStore) Spend(ctx context.Context, id uuid.UUID, amount int) (bool, error) {
	if amount <= 0 {
		return true, nil
	}
	res := s.db.WithContext(ctx).Model(&User{}).
		Where("id = ? AND credits >= ?", id, amount).
		UpdateColumn("credits", gorm.Expr("credits - ?", amount))
	if res.Error != nil {
		return false, res.Error
	}
	return res.RowsAffected == 1, nil
}

func (s *userStore) Refund(ctx context.Context, id uuid.UUID, amount int) error {
	if amount <= 0 {
		return nil
	}
	return s.db.WithContext(ctx).Model(&User{}).Where("id = ?", id).
		UpdateColumn("credits", gorm.Expr("credits + ?", amount)).Error
}

func IsNotFound(err error) bool { return errors.Is(err, gorm.ErrRecordNotFound) }
