package jwtx

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
)

type header struct {
	Alg string `json:"alg"`
	Kid string `json:"kid"`
}

// audience is a string in some tokens and an array in others.
type audience []string

func (a *audience) UnmarshalJSON(b []byte) error {
	var one string
	if err := json.Unmarshal(b, &one); err == nil {
		*a = audience{one}
		return nil
	}
	var many []string
	if err := json.Unmarshal(b, &many); err != nil {
		return err
	}
	*a = many
	return nil
}

func (a audience) has(want string) bool {
	for _, v := range a {
		if v == want {
			return true
		}
	}
	return false
}

type payload struct {
	Issuer    string   `json:"iss"`
	Subject   string   `json:"sub"`
	Audience  audience `json:"aud"`
	Expiry    int64    `json:"exp"`
	NotBefore int64    `json:"nbf"`
	Email     string   `json:"email"`
	Name      string   `json:"name"`
	Picture   string   `json:"picture"`
}

func split(token string) (header, payload, string, []byte, error) {
	var h header
	var p payload

	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return h, p, "", nil, errors.New("malformed token")
	}
	hb, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return h, p, "", nil, errors.New("malformed token header")
	}
	if err := json.Unmarshal(hb, &h); err != nil {
		return h, p, "", nil, errors.New("malformed token header")
	}
	pb, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return h, p, "", nil, errors.New("malformed token payload")
	}
	if err := json.Unmarshal(pb, &p); err != nil {
		return h, p, "", nil, errors.New("malformed token payload")
	}
	sig, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return h, p, "", nil, errors.New("malformed token signature")
	}
	return h, p, parts[0] + "." + parts[1], sig, nil
}

func verify(key *rsa.PublicKey, signingInput string, sig []byte) error {
	sum := sha256.Sum256([]byte(signingInput))
	if err := rsa.VerifyPKCS1v15(key, crypto.SHA256, sum[:], sig); err != nil {
		return errors.New("token signature is not valid")
	}
	return nil
}
