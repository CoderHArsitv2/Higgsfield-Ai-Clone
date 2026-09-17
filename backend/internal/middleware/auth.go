package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/coderHArsitv2/higgsfield-clone/backend/internal/models"
	"github.com/coderHArsitv2/higgsfield-clone/backend/pkg/apierr"
	"github.com/coderHArsitv2/higgsfield-clone/backend/pkg/jwtx"
)

const userKey = "current_user"

// Auth verifies the Auth0 access token and provisions the local user row on
// first sight. There is no signup endpoint by design: the first authenticated
// request *is* the signup, so Auth0 remains the only place identity is managed.
func Auth(v *jwtx.Validator, users models.UserStore) gin.HandlerFunc {
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

		user, err := users.Upsert(c.Request.Context(), *claims)
		if err != nil {
			apierr.Respond(c, err)
			return
		}

		c.Set(userKey, user)
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
