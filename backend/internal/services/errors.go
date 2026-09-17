package services

import (
	"errors"
	"net"
	"strings"

	"github.com/coderHArsitv2/higgsfield-clone/backend/pkg/httpx"
)

// retryable distinguishes an upstream blip worth another attempt from a
// permanent rejection. A 4xx means the request itself is wrong, so retrying it
// would only waste the user's quota.
func retryable(err error) bool {
	var se *httpx.StatusError
	if errors.As(err, &se) {
		return se.Status == 408 || se.Status == 429 || se.Status >= 500
	}
	var ne net.Error
	return errors.As(err, &ne)
}

// redact keeps provider error text useful without echoing a credential back to
// the user, since some vendors quote the offending key in their message.
func redact(err error) string {
	msg := err.Error()
	var se *httpx.StatusError
	if errors.As(err, &se) {
		switch {
		case se.Status == 401 || se.Status == 403:
			return "the API key for this provider was rejected"
		case se.Status == 429:
			return "provider rate limit reached, try again shortly"
		case se.Status >= 500:
			return "the provider is having problems right now"
		}
		msg = se.Body
	}
	if len(msg) > 400 {
		msg = msg[:400] + "…"
	}
	for _, marker := range []string{"sk-", "Bearer ", "key-", "fal_", "r8_"} {
		if i := strings.Index(msg, marker); i >= 0 {
			return "the provider rejected this request"
		}
	}
	return msg
}
