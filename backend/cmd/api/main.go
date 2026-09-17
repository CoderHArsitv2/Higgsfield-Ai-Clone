package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"github.com/coderHArsitv2/higgsfield-clone/backend/internal/config"
	"github.com/coderHArsitv2/higgsfield-clone/backend/internal/controllers"
	"github.com/coderHArsitv2/higgsfield-clone/backend/internal/models"
	"github.com/coderHArsitv2/higgsfield-clone/backend/internal/provider"
	"github.com/coderHArsitv2/higgsfield-clone/backend/internal/routes"
	"github.com/coderHArsitv2/higgsfield-clone/backend/internal/services"
	"github.com/coderHArsitv2/higgsfield-clone/backend/internal/storage"
	"github.com/coderHArsitv2/higgsfield-clone/backend/pkg/cryptox"
	"github.com/coderHArsitv2/higgsfield-clone/backend/pkg/jwtx"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(log)

	if err := run(log); err != nil {
		log.Error("fatal", "error", err)
		os.Exit(1)
	}
}

func run(log *slog.Logger) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	db, err := openDB(cfg, log)
	if err != nil {
		return err
	}
	// Explicit, reviewable SQL rather than AutoMigrate, which never drops or
	// narrows a column and so lets the schema drift silently from the structs.
	if err := models.Migrate(context.Background(), db, log); err != nil {
		return err
	}
	stores := models.New(db)

	cipher, err := cryptox.New(cfg.EncryptionKey)
	if err != nil {
		return err
	}

	// Object storage when a bucket is configured, local disk otherwise. Only
	// providers that hand back raw bytes rather than a URL -- OpenAI images,
	// Gemini, ElevenLabs -- ever touch this.
	mediaDir := envOr("MEDIA_DIR", "./.media")
	var store storage.Storage
	if cfg.Storage.Enabled() {
		store, err = storage.NewS3(storage.S3Config{
			Endpoint:      cfg.Storage.Endpoint,
			Region:        cfg.Storage.Region,
			Bucket:        cfg.Storage.Bucket,
			AccessKey:     cfg.Storage.AccessKey,
			SecretKey:     cfg.Storage.SecretKey,
			PublicBaseURL: cfg.Storage.PublicBaseURL,
			Prefix:        cfg.Storage.Prefix,
		})
		if err != nil {
			return err
		}
		log.Info("media storage: object store",
			"bucket", cfg.Storage.Bucket, "endpoint", cfg.Storage.Endpoint,
			"region", cfg.Storage.Region)
	} else {
		store, err = storage.NewLocal(mediaDir,
			envOr("PUBLIC_BASE_URL", "http://localhost:"+cfg.Port))
		if err != nil {
			return err
		}
		log.Warn("media storage: local disk — generated files are lost on restart",
			"dir", mediaDir, "hint", "set STORAGE_BUCKET to use an object store")
	}

	// Registration order is display order in the UI. Sandbox first so a new
	// install has something that works before any key is added.
	registry := provider.NewRegistry(cfg.ProviderKeys,
		provider.NewMock(),
		provider.NewFal(),
		provider.NewReplicate(),
		provider.NewOpenAI(store),
		provider.NewGoogle(store),
		provider.NewElevenLabs(store),
	)
	for _, p := range registry.Providers() {
		log.Info("provider registered",
			"id", p.ID(), "models", len(p.Models()),
			"unlocked", registry.PlatformCovers(p.ID()), "env_key", p.EnvKey())
	}

	keys := services.NewKeys(stores.Keys, cipher, registry)
	gens := services.NewGenerations(stores, registry, keys, log)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	services.NewWorker(stores, registry, keys, cfg.PollInterval, cfg.WorkerCount, log).Start(ctx)

	if cfg.IsProd() {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()
	routes.Register(r, routes.Options{
		Config:    cfg,
		Stores:    stores,
		Validator: jwtx.NewValidator(cfg.Auth0Domain, cfg.Auth0Audience),
		Deps: controllers.Deps{
			Registry: registry, Generations: gens, Keys: keys, Users: stores.Users,
		},
		MediaDir: mediaDir,
		Log:      log,
	})

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           r,
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		log.Info("listening", "port", cfg.Port, "env", cfg.Env,
			"version", cfg.Version, "commit", cfg.Commit)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("server error", "error", err)
			stop()
		}
	}()

	<-ctx.Done()
	log.Info("shutting down")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	return srv.Shutdown(shutdownCtx)
}

func openDB(cfg *config.Config, log *slog.Logger) (*gorm.DB, error) {
	level := gormlogger.Warn
	if !cfg.IsProd() {
		level = gormlogger.Error
	}
	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{
		Logger: gormlogger.Default.LogMode(level),
	})
	if err != nil {
		return nil, err
	}
	sql, err := db.DB()
	if err != nil {
		return nil, err
	}
	sql.SetMaxOpenConns(20)
	sql.SetMaxIdleConns(5)
	sql.SetConnMaxLifetime(time.Hour)
	return db, nil
}

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
