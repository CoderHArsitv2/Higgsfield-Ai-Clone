// Package storage re-hosts generated media that a provider hands back as raw
// bytes instead of a URL.
//
// Local disk is the default so the app runs with no cloud dependency. The
// interface is deliberately one method wide, so swapping in S3/R2 later is a
// constructor change and nothing else.
package storage

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Storage interface {
	Put(ctx context.Context, name string, data []byte, contentType string) (string, error)
}

type Local struct {
	dir     string
	baseURL string
}

func NewLocal(dir, baseURL string) (*Local, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("create media dir: %w", err)
	}
	return &Local{dir: dir, baseURL: strings.TrimSuffix(baseURL, "/")}, nil
}

func (l *Local) Put(_ context.Context, name string, data []byte, _ string) (string, error) {
	buf := make([]byte, 8)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	safe := hex.EncodeToString(buf) + "-" + filepath.Base(name)
	if err := os.WriteFile(filepath.Join(l.dir, safe), data, 0o644); err != nil {
		return "", err
	}
	return l.baseURL + "/media/" + safe, nil
}

func (l *Local) Dir() string { return l.dir }
