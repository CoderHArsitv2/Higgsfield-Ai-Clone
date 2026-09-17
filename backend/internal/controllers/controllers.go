// Package controllers holds the HTTP layer. Controllers translate requests into
// service calls and nothing more -- no provider calls, no SQL.
package controllers

import (
	"github.com/coderHArsitv2/higgsfield-clone/backend/internal/provider"
	"github.com/coderHArsitv2/higgsfield-clone/backend/internal/services"
)

type Deps struct {
	Registry    *provider.Registry
	Generations *services.Generations
	Keys        *services.Keys
}
