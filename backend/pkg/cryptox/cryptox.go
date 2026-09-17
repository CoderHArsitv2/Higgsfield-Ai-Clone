// Package cryptox encrypts user-supplied provider API keys at rest.
//
// A BYOK key is a live credential the user pays for, so it is AES-256-GCM
// encrypted with a server master key and is never returned to a client in
// plaintext -- only a masked preview is.
package cryptox

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
)

type Cipher struct{ aead cipher.AEAD }

// New takes a 64-character hex string (32 bytes).
func New(hexKey string) (*Cipher, error) {
	key, err := hex.DecodeString(hexKey)
	if err != nil {
		return nil, fmt.Errorf("encryption key is not valid hex: %w", err)
	}
	if len(key) != 32 {
		return nil, fmt.Errorf("encryption key must be 32 bytes, got %d", len(key))
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	return &Cipher{aead: aead}, nil
}

func (c *Cipher) Encrypt(plaintext string) (string, error) {
	nonce := make([]byte, c.aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	sealed := c.aead.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(sealed), nil
}

func (c *Cipher) Decrypt(encoded string) (string, error) {
	raw, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", err
	}
	if len(raw) < c.aead.NonceSize() {
		return "", errors.New("ciphertext too short")
	}
	nonce, body := raw[:c.aead.NonceSize()], raw[c.aead.NonceSize():]
	out, err := c.aead.Open(nil, nonce, body, nil)
	if err != nil {
		return "", errors.New("could not decrypt key")
	}
	return string(out), nil
}

// Mask renders a key for display: enough to recognise, useless if leaked.
func Mask(key string) string {
	if len(key) <= 8 {
		return "••••"
	}
	return key[:4] + "••••" + key[len(key)-4:]
}
