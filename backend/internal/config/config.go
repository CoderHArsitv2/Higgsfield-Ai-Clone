package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// Config is resolved once at boot. Provider credentials are deliberately kept
// in a map rather than named fields so adding a provider never requires
// touching this file -- the provider declares which env var it wants.
type Config struct {
	Env           string
	Version       string
	Commit        string
	Port          string
	DatabaseURL   string
	CORSOrigins   []string
	Auth0Domain   string
	Auth0Audience string
	EncryptionKey string
	WorkerCount   int
	PollInterval  time.Duration
	ProviderKeys  map[string]string
}

func Load() (*Config, error) {
	loadDotenv()

	c := &Config{
		Env: env("APP_ENV", "development"),
		// APP_VERSION is the release tag, set by the deploy workflow.
		// RENDER_GIT_COMMIT is injected by Render on every build, so even an
		// untagged deploy can be traced back to a commit.
		Version:       env("APP_VERSION", "dev"),
		Commit:        firstOf(os.Getenv("RENDER_GIT_COMMIT"), os.Getenv("GIT_COMMIT"), "unknown"),
		Port:          env("PORT", "8080"),
		DatabaseURL:   os.Getenv("DATABASE_URL"),
		Auth0Domain:   strings.TrimSuffix(os.Getenv("AUTH0_DOMAIN"), "/"),
		Auth0Audience: os.Getenv("AUTH0_AUDIENCE"),
		EncryptionKey: os.Getenv("ENCRYPTION_KEY"),
		ProviderKeys:  map[string]string{},
	}

	for _, o := range strings.Split(env("CORS_ORIGINS", "http://localhost:3000"), ",") {
		if o = strings.TrimSpace(o); o != "" {
			c.CORSOrigins = append(c.CORSOrigins, o)
		}
	}

	n, err := strconv.Atoi(env("WORKER_COUNT", "4"))
	if err != nil || n < 1 {
		n = 4
	}
	c.WorkerCount = n

	ms, err := strconv.Atoi(env("POLL_INTERVAL_MS", "3000"))
	if err != nil || ms < 500 {
		ms = 3000
	}
	c.PollInterval = time.Duration(ms) * time.Millisecond

	// Every provider credential in the environment, collected by convention.
	for _, kv := range os.Environ() {
		parts := strings.SplitN(kv, "=", 2)
		if len(parts) == 2 && parts[1] != "" {
			c.ProviderKeys[parts[0]] = parts[1]
		}
	}

	return c, c.validate()
}

// loadDotenv walks up from the working directory looking for a .env file.
//
// Searching upward rather than using a fixed relative path means it works from
// wherever the process was started -- `go run ./cmd/api` from backend/, `go run
// main.go` from backend/cmd/api/, or the compiled binary from anywhere.
//
// godotenv does not overwrite variables that are already set, so a real
// deployment (Render, CI) always wins over a file that happens to be present.
// A missing .env is not an error: production has none.
func loadDotenv() {
	dir, err := os.Getwd()
	if err != nil {
		return
	}
	for i := 0; i < 6; i++ {
		candidate := filepath.Join(dir, ".env")
		if _, err := os.Stat(candidate); err == nil {
			_ = godotenv.Load(candidate)
			return
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return // reached the filesystem root
		}
		dir = parent
	}
}

func (c *Config) validate() error {
	if c.DatabaseURL == "" {
		return fmt.Errorf("DATABASE_URL is required")
	}
	if c.Auth0Domain == "" || c.Auth0Audience == "" {
		return fmt.Errorf("AUTH0_DOMAIN and AUTH0_AUDIENCE are required")
	}
	// A short key would silently weaken BYOK encryption, so fail loudly at boot
	// rather than at the moment a user saves their first key.
	if len(c.EncryptionKey) != 64 {
		return fmt.Errorf("ENCRYPTION_KEY must be 64 hex characters (32 bytes); generate with: openssl rand -hex 32")
	}
	return nil
}

func (c *Config) IsProd() bool { return c.Env == "production" }

func firstOf(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

func env(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
