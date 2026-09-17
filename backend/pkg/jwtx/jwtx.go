// Package jwtx validates Auth0-issued access tokens.
//
// Auth0 is the only identity source in this app, so the backend stores no
// passwords and holds no sessions: it verifies the RS256 signature against the
// tenant's published JWKS and trusts the claims. Keys are cached and refetched
// only when an unknown key id appears (a rotation), which keeps the hot path
// free of network calls without pinning a stale key forever.
package jwtx

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"sync"
	"time"
)

type Claims struct {
	Subject string
	Email   string
	Name    string
	Picture string
}

type Validator struct {
	issuer   string
	audience string
	jwksURL  string
	client   *http.Client

	mu        sync.RWMutex
	keys      map[string]*rsa.PublicKey
	lastFetch time.Time
}

func NewValidator(domain, audience string) *Validator {
	issuer := "https://" + domain + "/"
	return &Validator{
		issuer:   issuer,
		audience: audience,
		jwksURL:  issuer + ".well-known/jwks.json",
		client:   &http.Client{Timeout: 10 * time.Second},
		keys:     map[string]*rsa.PublicKey{},
	}
}

func (v *Validator) Validate(ctx context.Context, token string) (*Claims, error) {
	header, payload, signingInput, sig, err := split(token)
	if err != nil {
		return nil, err
	}
	if header.Alg != "RS256" {
		return nil, fmt.Errorf("unsupported signing algorithm %q", header.Alg)
	}

	key, err := v.key(ctx, header.Kid)
	if err != nil {
		return nil, err
	}
	if err := verify(key, signingInput, sig); err != nil {
		return nil, err
	}

	if payload.Issuer != v.issuer {
		return nil, fmt.Errorf("unexpected issuer %q", payload.Issuer)
	}
	if !payload.Audience.has(v.audience) {
		return nil, errors.New("token audience does not match this API")
	}
	now := time.Now().Unix()
	if payload.Expiry != 0 && now >= payload.Expiry {
		return nil, errors.New("token has expired")
	}
	if payload.NotBefore != 0 && now < payload.NotBefore {
		return nil, errors.New("token is not valid yet")
	}
	if payload.Subject == "" {
		return nil, errors.New("token has no subject")
	}

	return &Claims{
		Subject: payload.Subject,
		Email:   payload.Email,
		Name:    payload.Name,
		Picture: payload.Picture,
	}, nil
}

func (v *Validator) key(ctx context.Context, kid string) (*rsa.PublicKey, error) {
	v.mu.RLock()
	k, ok := v.keys[kid]
	v.mu.RUnlock()
	if ok {
		return k, nil
	}

	// Unknown kid: either first use or a key rotation. Refetch, but not more
	// than once a minute, so an invalid kid cannot be used to hammer Auth0.
	v.mu.Lock()
	defer v.mu.Unlock()
	if k, ok := v.keys[kid]; ok {
		return k, nil
	}
	if time.Since(v.lastFetch) < time.Minute && len(v.keys) > 0 {
		return nil, fmt.Errorf("unknown signing key %q", kid)
	}
	if err := v.refresh(ctx); err != nil {
		return nil, err
	}
	if k, ok := v.keys[kid]; ok {
		return k, nil
	}
	return nil, fmt.Errorf("unknown signing key %q", kid)
}

func (v *Validator) refresh(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, v.jwksURL, nil)
	if err != nil {
		return err
	}
	res, err := v.client.Do(req)
	if err != nil {
		return fmt.Errorf("fetch jwks: %w", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("jwks endpoint returned %d", res.StatusCode)
	}

	var doc struct {
		Keys []struct {
			Kid string `json:"kid"`
			Kty string `json:"kty"`
			N   string `json:"n"`
			E   string `json:"e"`
		} `json:"keys"`
	}
	if err := json.NewDecoder(res.Body).Decode(&doc); err != nil {
		return err
	}

	fresh := map[string]*rsa.PublicKey{}
	for _, k := range doc.Keys {
		if k.Kty != "RSA" {
			continue
		}
		nb, err := base64.RawURLEncoding.DecodeString(k.N)
		if err != nil {
			continue
		}
		eb, err := base64.RawURLEncoding.DecodeString(k.E)
		if err != nil {
			continue
		}
		fresh[k.Kid] = &rsa.PublicKey{
			N: new(big.Int).SetBytes(nb),
			E: int(new(big.Int).SetBytes(eb).Int64()),
		}
	}
	if len(fresh) == 0 {
		return errors.New("jwks contained no usable RSA keys")
	}
	v.keys = fresh
	v.lastFetch = time.Now()
	return nil
}
