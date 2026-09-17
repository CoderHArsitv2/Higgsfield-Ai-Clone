package middleware

import (
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/coderHArsitv2/higgsfield-clone/backend/internal/models"
	"github.com/coderHArsitv2/higgsfield-clone/backend/pkg/apierr"
	"github.com/coderHArsitv2/higgsfield-clone/backend/pkg/jwtx"
)

const userKey = "current_user"

// Auth verifies the Auth0 access token and provisions the local user row on
// first sight. There is no signup endpoint by design: the first authenticated
// request *is* the signup, so Auth0 remains the only place identity is managed.
func Auth(v *jwtx.Validator, db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		raw := c.GetHeader("Authorization")
		if !strings.HasPrefix(strings.ToLower(raw), "bearer ") {
			apierr.Respond(c, apierr.ErrUnauthorized)
			return
		}

		claims, err := v.Validate(c.Request.Context(), strings.TrimSpace(raw[7:]))
		if err != nil {
			apierr.Respond(c, apierr.New(401, "invalid_token", err.Error()))
			return
		}

		now := time.Now()
		user := models.User{
			Auth0Subject: claims.Subject,
			Email:        claims.Email,
			Name:         claims.Name,
			Picture:      claims.Picture,
			LastSeenAt:   &now,
		}
		// Profile fields are refreshed from the token; credits are not, so a
		// re-login can never top someone up.
		err = db.WithContext(c.Request.Context()).Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "auth0_subject"}},
			DoUpdates: clause.AssignmentColumns([]string{"email", "name", "picture", "last_seen_at", "updated_at"}),
		}).Create(&user).Error
		if err != nil {
			apierr.Respond(c, err)
			return
		}

		c.Set(userKey, &user)
		c.Next()
	}
}

// CurrentUser is never called outside an Auth-protected route, so a miss here
// is a routing bug rather than an auth failure.
func CurrentUser(c *gin.Context) *models.User {
	v, ok := c.Get(userKey)
	if !ok {
		return nil
	}
	u, _ := v.(*models.User)
	return u
}
