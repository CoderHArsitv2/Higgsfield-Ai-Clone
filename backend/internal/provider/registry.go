package provider

import (
	"sort"
	"strings"
)

// Access says why a model is or is not usable by a given user. The frontend
// renders locked models rather than hiding them -- seeing the full catalogue
// is the point, and a lock is an invitation to add a key.
type Access string

const (
	AccessServer Access = "server" // platform credential covers it
	AccessBYOK   Access = "byok"   // this user supplied their own key
	AccessLocked Access = "locked" // no credential anywhere
)

type ModelView struct {
	ModelSpec
	ProviderName string `json:"provider_name"`
	Access       Access `json:"access"`
	Enabled      bool   `json:"enabled"`
	LockReason   string `json:"lock_reason,omitempty"`
	KeyEnvName   string `json:"key_env_name,omitempty"`
	DocsURL      string `json:"docs_url,omitempty"`
}

type ProviderView struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	EnvKey        string `json:"env_key"`
	DocsURL       string `json:"docs_url"`
	HasServerKey  bool   `json:"has_server_key"`
	AcceptsBYOK   bool   `json:"accepts_byok"`
	ModelCount    int    `json:"model_count"`
	EnabledModels int    `json:"enabled_models"`
}

type Registry struct {
	order      []Provider
	byID       map[string]Provider
	serverKeys map[string]string
	keyless    map[string]bool
	modelIndex map[string]indexed
}

type indexed struct {
	provider Provider
	spec     ModelSpec
}

// NewRegistry wires the providers and resolves which ones the server itself can
// pay for. env is the full process environment.
func NewRegistry(env map[string]string, providers ...Provider) *Registry {
	r := &Registry{
		byID:       map[string]Provider{},
		serverKeys: map[string]string{},
		keyless:    map[string]bool{},
		modelIndex: map[string]indexed{},
	}
	for _, p := range providers {
		r.order = append(r.order, p)
		r.byID[p.ID()] = p
		// A provider that declares no env key needs no credential at all (the
		// sandbox). Without this it would look up env[""], find nothing, and
		// render as locked -- which would defeat the point of having it.
		if p.EnvKey() == "" {
			r.keyless[p.ID()] = true
		} else if key := strings.TrimSpace(env[p.EnvKey()]); key != "" {
			r.serverKeys[p.ID()] = key
		}
		for _, m := range p.Models() {
			m.ProviderID = p.ID()
			r.modelIndex[m.ID] = indexed{provider: p, spec: m}
		}
	}
	return r
}

func (r *Registry) Providers() []Provider { return r.order }

func (r *Registry) Provider(id string) (Provider, bool) {
	p, ok := r.byID[id]
	return p, ok
}

func (r *Registry) Model(modelID string) (Provider, ModelSpec, bool) {
	e, ok := r.modelIndex[modelID]
	if !ok {
		return nil, ModelSpec{}, false
	}
	return e.provider, e.spec, true
}

func (r *Registry) ServerKey(providerID string) (string, bool) {
	k, ok := r.serverKeys[providerID]
	return k, ok
}

// PlatformCovers reports whether the platform can run this provider without the
// user supplying anything -- either it needs no key, or the server holds one.
func (r *Registry) PlatformCovers(providerID string) bool {
	if r.keyless[providerID] {
		return true
	}
	_, ok := r.serverKeys[providerID]
	return ok
}

// ResolveKey picks the credential for a call. A user's own key always wins, so
// BYOK also acts as an escape hatch from platform rate limits.
func (r *Registry) ResolveKey(providerID string, userKey string) (key string, ownKey bool, err error) {
	if r.keyless[providerID] {
		return "", false, nil
	}
	if strings.TrimSpace(userKey) != "" {
		return userKey, true, nil
	}
	if k, ok := r.serverKeys[providerID]; ok {
		return k, false, nil
	}
	return "", false, ErrMissingKey
}

// Catalog returns every model in the system, annotated for this user.
// userProviders is the set of provider ids the user has a stored key for.
func (r *Registry) Catalog(userProviders map[string]bool) []ModelView {
	var out []ModelView
	for _, p := range r.order {
		hasServer := r.PlatformCovers(p.ID())
		hasUser := userProviders[p.ID()]
		for _, m := range p.Models() {
			m.ProviderID = p.ID()
			v := ModelView{
				ModelSpec:    m,
				ProviderName: p.Name(),
				KeyEnvName:   p.EnvKey(),
				DocsURL:      p.DocsURL(),
			}
			switch {
			case hasServer:
				v.Access, v.Enabled = AccessServer, true
			case hasUser:
				v.Access, v.Enabled = AccessBYOK, true
			default:
				v.Access, v.Enabled = AccessLocked, false
				v.LockReason = "Add a " + p.Name() + " API key to unlock"
			}
			out = append(out, v)
		}
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Enabled != out[j].Enabled {
			return out[i].Enabled
		}
		if out[i].Featured != out[j].Featured {
			return out[i].Featured
		}
		return out[i].Name < out[j].Name
	})
	return out
}

func (r *Registry) ProviderViews(userProviders map[string]bool) []ProviderView {
	var out []ProviderView
	for _, p := range r.order {
		hasServer := r.PlatformCovers(p.ID())
		enabled := 0
		if hasServer || userProviders[p.ID()] {
			enabled = len(p.Models())
		}
		out = append(out, ProviderView{
			ID:            p.ID(),
			Name:          p.Name(),
			EnvKey:        p.EnvKey(),
			DocsURL:       p.DocsURL(),
			HasServerKey:  hasServer,
			AcceptsBYOK:   p.EnvKey() != "",
			ModelCount:    len(p.Models()),
			EnabledModels: enabled,
		})
	}
	return out
}
