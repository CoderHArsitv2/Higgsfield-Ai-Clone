package apierr

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Error struct {
	Status  int    `json:"-"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e *Error) Error() string { return e.Message }

func New(status int, code, msg string) *Error {
	return &Error{Status: status, Code: code, Message: msg}
}

var (
	ErrUnauthorized = New(http.StatusUnauthorized, "unauthorized", "authentication required")
	ErrForbidden    = New(http.StatusForbidden, "forbidden", "not permitted")
	ErrNotFound     = New(http.StatusNotFound, "not_found", "resource not found")
	ErrInternal     = New(http.StatusInternalServerError, "internal_error", "something went wrong")
)

func BadRequest(msg string) *Error  { return New(http.StatusBadRequest, "bad_request", msg) }
func Conflict(msg string) *Error    { return New(http.StatusConflict, "conflict", msg) }
func Unprocessable(m string) *Error { return New(http.StatusUnprocessableEntity, "unprocessable", m) }

// Respond writes an error to the client. Unknown errors are never echoed back
// verbatim -- a provider error can contain an API key.
func Respond(c *gin.Context, err error) {
	if e, ok := err.(*Error); ok {
		c.AbortWithStatusJSON(e.Status, gin.H{"error": e})
		return
	}
	_ = c.Error(err)
	c.AbortWithStatusJSON(ErrInternal.Status, gin.H{"error": ErrInternal})
}
