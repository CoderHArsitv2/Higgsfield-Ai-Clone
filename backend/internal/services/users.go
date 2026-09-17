package services

import (
	"context"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/coderHArsitv2/higgsfield-clone/backend/internal/models"
	"github.com/coderHArsitv2/higgsfield-clone/backend/pkg/jwtx"
)

type Users struct{ db *gorm.DB }

func NewUsers(db *gorm.DB) *Users { return &Users{db: db} }

// Upsert provisions the local row for an Auth0 identity and returns it as the
// database actually holds it.
//
// The RETURNING clause is load-bearing. BeforeCreate mints a fresh UUID on
// every request; on a conflict Postgres updates the existing row and keeps its
// original id, but without RETURNING, GORM leaves the struct holding the new
// UUID that was never written. Every request after the first login then carried
// a user id that does not exist -- BYOK lookups silently found nothing, and
// creating a generation failed on the users foreign key.
func (u *Users) Upsert(ctx context.Context, claims jwtx.Claims) (*models.User, error) {
	now := time.Now()
	row := models.User{
		Auth0Subject: claims.Subject,
		Email:        claims.Email,
		Name:         claims.Name,
		Picture:      claims.Picture,
		LastSeenAt:   &now,
	}

	// Profile fields are refreshed from the token; credits are not, so a
	// re-login can never top someone up.
	err := u.db.WithContext(ctx).Clauses(
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
