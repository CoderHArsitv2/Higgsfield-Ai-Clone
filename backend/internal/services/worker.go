package services

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"sync"
	"time"

	"gorm.io/gorm"

	"github.com/coderHArsitv2/higgsfield-clone/backend/internal/models"
	"github.com/coderHArsitv2/higgsfield-clone/backend/internal/provider"
)

// Worker drives every generation from queued to terminal.
//
// It is a database-backed queue rather than an in-memory one on purpose: a
// deploy or crash mid-generation must not lose a job the user already paid
// credits for. On restart the same rows are picked up exactly where they were.
type Worker struct {
	db       *gorm.DB
	registry *provider.Registry
	keys     *Keys
	interval time.Duration
	parallel int
	log      *slog.Logger

	mu       sync.Mutex
	inFlight map[string]bool
}

const (
	maxAttempts = 3
	jobTimeout  = 20 * time.Minute
)

func NewWorker(db *gorm.DB, reg *provider.Registry, keys *Keys, interval time.Duration, parallel int, log *slog.Logger) *Worker {
	return &Worker{
		db: db, registry: reg, keys: keys,
		interval: interval, parallel: parallel, log: log,
		inFlight: map[string]bool{},
	}
}

func (w *Worker) Start(ctx context.Context) {
	go func() {
		t := time.NewTicker(w.interval)
		defer t.Stop()
		w.log.Info("generation worker started", "interval", w.interval, "parallel", w.parallel)
		for {
			select {
			case <-ctx.Done():
				w.log.Info("generation worker stopped")
				return
			case <-t.C:
				w.tick(ctx)
			}
		}
	}()
}

func (w *Worker) tick(ctx context.Context) {
	var jobs []models.Generation
	err := w.db.WithContext(ctx).
		Where("status IN ?", []models.GenerationStatus{models.StatusQueued, models.StatusRunning}).
		Order("created_at ASC").Limit(w.parallel * 4).Find(&jobs).Error
	if err != nil {
		w.log.Error("worker could not load jobs", "error", err)
		return
	}

	sem := make(chan struct{}, w.parallel)
	var wg sync.WaitGroup
	for i := range jobs {
		job := jobs[i]
		if !w.claim(job.ID.String()) {
			continue
		}
		wg.Add(1)
		sem <- struct{}{}
		go func() {
			defer wg.Done()
			defer func() { <-sem }()
			defer w.release(job.ID.String())
			if err := w.process(ctx, &job); err != nil {
				w.log.Error("job failed", "generation", job.ID, "model", job.ModelID, "error", err)
			}
		}()
	}
	wg.Wait()
}

func (w *Worker) claim(id string) bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.inFlight[id] {
		return false
	}
	w.inFlight[id] = true
	return true
}

func (w *Worker) release(id string) {
	w.mu.Lock()
	delete(w.inFlight, id)
	w.mu.Unlock()
}

func (w *Worker) process(ctx context.Context, gen *models.Generation) error {
	// A job that has been running far too long is almost certainly an upstream
	// black hole. Fail it so the user gets their credits back.
	if gen.StartedAt != nil && time.Since(*gen.StartedAt) > jobTimeout {
		return w.fail(ctx, gen, "generation timed out upstream")
	}

	p, spec, ok := w.registry.Model(gen.ModelID)
	if !ok {
		return w.fail(ctx, gen, "model "+gen.ModelID+" is no longer available")
	}

	key, err := w.credential(ctx, gen, p.ID())
	if err != nil {
		return w.fail(ctx, gen, err.Error())
	}

	if gen.Status == models.StatusQueued {
		return w.submit(ctx, gen, p, spec, key)
	}
	return w.poll(ctx, gen, p, spec, key)
}

func (w *Worker) credential(ctx context.Context, gen *models.Generation, providerID string) (string, error) {
	var userKey string
	if gen.UsedOwnKey {
		var err error
		userKey, err = w.keys.Plaintext(ctx, gen.UserID, providerID)
		if err != nil {
			return "", err
		}
		if userKey == "" {
			return "", errors.New("your API key for this provider was removed before the job ran")
		}
	}
	key, _, err := w.registry.ResolveKey(providerID, userKey)
	if err != nil {
		return "", errors.New("no API key available for this provider")
	}
	return key, nil
}

func (w *Worker) submit(ctx context.Context, gen *models.Generation, p provider.Provider, spec provider.ModelSpec, key string) error {
	// Claim the row in the database too, so a second instance cannot submit the
	// same job twice and bill the user twice.
	now := time.Now()
	res := w.db.WithContext(ctx).Model(&models.Generation{}).
		Where("id = ? AND status = ?", gen.ID, models.StatusQueued).
		Updates(map[string]any{
			"status":     models.StatusRunning,
			"started_at": &now,
			"attempts":   gorm.Expr("attempts + 1"),
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return nil // another worker got there first
	}
	gen.Status = models.StatusRunning
	gen.StartedAt = &now

	var params map[string]any
	_ = json.Unmarshal(gen.Params, &params)

	out, err := p.Submit(ctx, provider.SubmitRequest{
		Model:  spec,
		Prompt: gen.Prompt,
		Params: params,
		APIKey: key,
	})
	if err != nil {
		if gen.Attempts+1 < maxAttempts && retryable(err) {
			// Put it back in the queue for another pass rather than burning the
			// user's credits on a transient upstream blip.
			return w.db.WithContext(ctx).Model(gen).
				Updates(map[string]any{"status": models.StatusQueued, "started_at": nil}).Error
		}
		return w.fail(ctx, gen, redact(err))
	}

	if out.Done {
		return w.succeed(ctx, gen, out.Assets)
	}
	return w.db.WithContext(ctx).Model(gen).UpdateColumn("external_id", out.ExternalID).Error
}

func (w *Worker) poll(ctx context.Context, gen *models.Generation, p provider.Provider, spec provider.ModelSpec, key string) error {
	if gen.ExternalID == "" {
		return nil // submitted but the handle has not been written yet
	}
	out, err := p.Poll(ctx, provider.PollRequest{Model: spec, ExternalID: gen.ExternalID, APIKey: key})
	if err != nil {
		w.log.Warn("poll failed, will retry", "generation", gen.ID, "error", err)
		return nil
	}
	switch {
	case out.Failed:
		return w.fail(ctx, gen, out.Error)
	case out.Done:
		return w.succeed(ctx, gen, out.Assets)
	}
	return nil
}

func (w *Worker) succeed(ctx context.Context, gen *models.Generation, assets []provider.ResultAsset) error {
	now := time.Now()
	return w.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, a := range assets {
			row := models.Asset{
				GenerationID: gen.ID,
				Kind:         a.Kind,
				URL:          a.URL,
				ThumbnailURL: a.ThumbnailURL,
				Width:        a.Width,
				Height:       a.Height,
				DurationMS:   a.DurationMS,
			}
			if err := tx.Create(&row).Error; err != nil {
				return err
			}
		}
		return tx.Model(&models.Generation{}).Where("id = ?", gen.ID).
			Updates(map[string]any{
				"status":       models.StatusSucceeded,
				"completed_at": &now,
			}).Error
	})
}

// fail marks the job failed and refunds credits, because the user got nothing.
func (w *Worker) fail(ctx context.Context, gen *models.Generation, reason string) error {
	if reason == "" {
		reason = "generation failed"
	}
	now := time.Now()
	return w.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		res := tx.Model(&models.Generation{}).
			Where("id = ? AND status NOT IN ?", gen.ID,
				[]models.GenerationStatus{models.StatusSucceeded, models.StatusFailed, models.StatusCanceled}).
			Updates(map[string]any{
				"status":       models.StatusFailed,
				"error":        reason,
				"completed_at": &now,
			})
		if res.Error != nil {
			return res.Error
		}
		// Only refund if this call is the one that transitioned the row --
		// otherwise a double poll would refund twice.
		if res.RowsAffected == 0 || gen.UsedOwnKey {
			return nil
		}
		if _, spec, ok := w.registry.Model(gen.ModelID); ok && spec.CreditCost > 0 {
			return tx.Model(&models.User{}).Where("id = ?", gen.UserID).
				UpdateColumn("credits", gorm.Expr("credits + ?", spec.CreditCost)).Error
		}
		return nil
	})
}
