package services

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"sync"
	"time"

	"github.com/coderHArsitv2/higgsfield-clone/backend/internal/models"
	"github.com/coderHArsitv2/higgsfield-clone/backend/internal/provider"
)

// Worker drives every generation from queued to terminal.
//
// It is a database-backed queue rather than an in-memory one on purpose: a
// deploy or crash mid-generation must not lose a job the user already paid
// credits for. On restart the same rows are picked up exactly where they were.
//
// Every provider call is logged with its outcome and duration, so "did a
// request actually go out, and what came back" is answerable from the log alone
// rather than by adding instrumentation after the fact.
type Worker struct {
	stores   *models.Stores
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

func NewWorker(stores *models.Stores, reg *provider.Registry, keys *Keys, interval time.Duration, parallel int, log *slog.Logger) *Worker {
	return &Worker{
		stores: stores, registry: reg, keys: keys,
		interval: interval, parallel: parallel,
		log:      log.With("component", "generation-worker"),
		inFlight: map[string]bool{},
	}
}

func (w *Worker) Start(ctx context.Context) {
	go func() {
		t := time.NewTicker(w.interval)
		defer t.Stop()
		w.log.Info("worker started", "interval", w.interval, "parallel", w.parallel)
		for {
			select {
			case <-ctx.Done():
				w.log.Info("worker stopped")
				return
			case <-t.C:
				w.tick(ctx)
			}
		}
	}()
}

func (w *Worker) tick(ctx context.Context) {
	jobs, err := w.stores.Generations.Pending(ctx, w.parallel*4)
	if err != nil {
		w.log.Error("could not load pending jobs", "error", err)
		return
	}
	if len(jobs) == 0 {
		return
	}
	w.log.Debug("pending jobs", "count", len(jobs))

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
				w.log.Error("job processing failed",
					"generation", job.ID, "model", job.ModelID, "error", err)
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
	log := w.log.With(
		"generation", gen.ID, "user", gen.UserID,
		"model", gen.ModelID, "provider", gen.ProviderID, "modality", gen.Modality)

	// A job running far too long is almost certainly an upstream black hole.
	// Fail it so the user gets their credits back.
	if gen.StartedAt != nil && time.Since(*gen.StartedAt) > jobTimeout {
		log.Warn("job timed out upstream", "running_for", time.Since(*gen.StartedAt))
		return w.fail(ctx, gen, "generation timed out upstream")
	}

	p, spec, ok := w.registry.Model(gen.ModelID)
	if !ok {
		log.Error("model no longer registered")
		return w.fail(ctx, gen, "model "+gen.ModelID+" is no longer available")
	}

	key, err := w.credential(ctx, gen, p.ID())
	if err != nil {
		log.Warn("no usable credential", "error", err)
		return w.fail(ctx, gen, err.Error())
	}

	if gen.Status == models.StatusQueued {
		return w.submit(ctx, log, gen, p, spec, key)
	}
	return w.poll(ctx, log, gen, p, spec, key)
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

func (w *Worker) submit(ctx context.Context, log *slog.Logger, gen *models.Generation, p provider.Provider, spec provider.ModelSpec, key string) error {
	// Claim the row in the database too, so a second instance cannot submit the
	// same job twice and bill the user twice.
	claimed, err := w.stores.Generations.ClaimForSubmit(ctx, gen.ID)
	if err != nil {
		return err
	}
	if !claimed {
		return nil // another worker got there first
	}
	gen.Status = models.StatusRunning

	var params map[string]any
	_ = json.Unmarshal(gen.Params, &params)

	log.Info("submitting to provider", "attempt", gen.Attempts+1, "params", params)
	started := time.Now()

	out, err := p.Submit(ctx, provider.SubmitRequest{
		Model:  spec,
		Prompt: gen.Prompt,
		Params: params,
		APIKey: key,
	})
	took := time.Since(started).Round(time.Millisecond)

	if err != nil {
		retry := gen.Attempts+1 < maxAttempts && retryable(err)
		log.Error("provider submit failed",
			"took", took, "retryable", retry, "error", redact(err))
		if retry {
			// Back into the queue rather than burning the user's credits on a
			// transient upstream blip.
			return w.stores.Generations.Requeue(ctx, gen.ID)
		}
		return w.fail(ctx, gen, redact(err))
	}

	if out.Done {
		log.Info("provider returned synchronously", "took", took, "assets", len(out.Assets))
		return w.succeed(ctx, log, gen, out.Assets)
	}

	log.Info("provider accepted job", "took", took, "external_id", out.ExternalID)
	return w.stores.Generations.SetExternalID(ctx, gen.ID, out.ExternalID)
}

func (w *Worker) poll(ctx context.Context, log *slog.Logger, gen *models.Generation, p provider.Provider, spec provider.ModelSpec, key string) error {
	if gen.ExternalID == "" {
		return nil // submitted but the handle has not been written yet
	}

	started := time.Now()
	out, err := p.Poll(ctx, provider.PollRequest{
		Model: spec, ExternalID: gen.ExternalID, APIKey: key,
	})
	took := time.Since(started).Round(time.Millisecond)

	if err != nil {
		// A dropped poll is not fatal; the next tick retries.
		log.Warn("poll failed, will retry", "took", took, "error", redact(err))
		return nil
	}

	switch {
	case out.Failed:
		log.Error("provider reported failure", "took", took, "reason", out.Error)
		return w.fail(ctx, gen, out.Error)
	case out.Done:
		log.Info("generation complete", "took", took, "assets", len(out.Assets),
			"total", time.Since(gen.CreatedAt).Round(time.Second))
		return w.succeed(ctx, log, gen, out.Assets)
	default:
		log.Debug("still running", "took", took, "progress", out.Progress)
		return nil
	}
}

func (w *Worker) succeed(ctx context.Context, log *slog.Logger, gen *models.Generation, assets []provider.ResultAsset) error {
	rows := make([]models.Asset, 0, len(assets))
	for _, a := range assets {
		rows = append(rows, models.Asset{
			GenerationID: gen.ID,
			Kind:         a.Kind,
			URL:          a.URL,
			ThumbnailURL: a.ThumbnailURL,
			Width:        a.Width,
			Height:       a.Height,
			DurationMS:   a.DurationMS,
		})
	}

	err := w.stores.Tx(ctx, func(tx *models.Stores) error {
		if err := tx.Assets.CreateMany(ctx, rows); err != nil {
			return err
		}
		_, err := tx.Generations.Finish(ctx, gen.ID, models.StatusSucceeded, "")
		return err
	})
	if err != nil {
		log.Error("could not record success", "error", err)
		return err
	}
	log.Info("generation succeeded", "assets", len(rows))
	return nil
}

// fail marks the job failed and refunds credits, because the user got nothing.
func (w *Worker) fail(ctx context.Context, gen *models.Generation, reason string) error {
	if reason == "" {
		reason = "generation failed"
	}
	return w.stores.Tx(ctx, func(tx *models.Stores) error {
		transitioned, err := tx.Generations.Finish(ctx, gen.ID, models.StatusFailed, reason)
		if err != nil {
			return err
		}
		// Only the call that performed the transition may refund, or a double
		// poll refunds twice.
		if !transitioned || gen.UsedOwnKey {
			return nil
		}
		_, spec, ok := w.registry.Model(gen.ModelID)
		if !ok || spec.CreditCost == 0 {
			return nil
		}
		if err := tx.Users.Refund(ctx, gen.UserID, spec.CreditCost); err != nil {
			return err
		}
		w.log.Info("credits refunded",
			"generation", gen.ID, "user", gen.UserID, "amount", spec.CreditCost)
		return nil
	})
}
