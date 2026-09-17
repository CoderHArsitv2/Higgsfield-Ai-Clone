package services

import (
	"context"
	"io"
	"log/slog"
	"os"
	"testing"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"github.com/coderHArsitv2/higgsfield-clone/backend/internal/models"
	"github.com/coderHArsitv2/higgsfield-clone/backend/internal/provider"
	"github.com/coderHArsitv2/higgsfield-clone/backend/pkg/cryptox"
	"github.com/coderHArsitv2/higgsfield-clone/backend/pkg/jwtx"
)

const testEncryptionKey = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

// Exercises the whole generation loop against a real database: create a job,
// let the worker submit and poll it, and check the result and the credit
// accounting. Skipped unless TEST_DATABASE_URL is set, so `go test ./...` stays
// runnable with no infrastructure.
func testDB(t *testing.T) *gorm.DB {
	t.Helper()
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL not set; skipping integration test")
	}
	db, err := gorm.Open(postgres.Open(url), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	if err := db.AutoMigrate(models.All()...); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

func newUser(t *testing.T, db *gorm.DB, credits int) *models.User {
	t.Helper()
	u := models.User{
		Auth0Subject: "test|" + time.Now().Format("150405.000000000"),
		Email:        "test@example.com",
	}
	if err := db.Create(&u).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	// Credits carries a `default:100` tag, and GORM omits zero-valued fields
	// with a default from the INSERT so the database default wins. Setting the
	// balance has to be a separate write, or a "user with 0 credits" silently
	// becomes a user with 100 and the test asserts nothing.
	if err := db.Model(&u).UpdateColumn("credits", credits).Error; err != nil {
		t.Fatalf("set credits: %v", err)
	}
	u.Credits = credits
	t.Cleanup(func() { db.Unscoped().Delete(&u) })
	return &u
}

func harness(t *testing.T, db *gorm.DB) (*Generations, *Worker) {
	t.Helper()
	reg := provider.NewRegistry(map[string]string{}, provider.NewMock(), provider.NewFal())
	cipher, err := cryptox.New(testEncryptionKey)
	if err != nil {
		t.Fatal(err)
	}
	keys := NewKeys(db, cipher, reg)
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	return NewGenerations(db, reg, keys), NewWorker(db, reg, keys, time.Second, 2, log)
}

func TestGenerationReachesSuccess(t *testing.T) {
	db := testDB(t)
	user := newUser(t, db, 100)
	gens, worker := harness(t, db)
	ctx := context.Background()

	gen, err := gens.Create(ctx, user, CreateInput{
		ModelID: "mock/still",
		Prompt:  "a rain-soaked alley at night",
		Params:  map[string]any{"aspect_ratio": "16:9"},
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if gen.Status != models.StatusQueued {
		t.Fatalf("expected queued, got %q", gen.Status)
	}

	// The sandbox model costs 1 credit and the user did not bring a key.
	var after models.User
	db.First(&after, "id = ?", user.ID)
	if after.Credits != 99 {
		t.Fatalf("expected 99 credits after a 1-credit job, got %d", after.Credits)
	}

	deadline := time.Now().Add(30 * time.Second)
	for time.Now().Before(deadline) {
		worker.tick(ctx)
		final, err := gens.Get(ctx, user.ID, gen.ID)
		if err != nil {
			t.Fatalf("get: %v", err)
		}
		if final.Status.Terminal() {
			if final.Status != models.StatusSucceeded {
				t.Fatalf("job ended %q: %s", final.Status, final.Error)
			}
			if len(final.Assets) == 0 {
				t.Fatal("succeeded with no assets")
			}
			for _, a := range final.Assets {
				if a.URL == "" {
					t.Fatal("asset has no url")
				}
			}
			return
		}
		time.Sleep(500 * time.Millisecond)
	}
	t.Fatal("job did not finish within 30s")
}

// A locked model must be refused at creation time, before any credits move.
func TestCreateRejectsLockedModel(t *testing.T) {
	db := testDB(t)
	user := newUser(t, db, 100)
	gens, _ := harness(t, db)

	_, err := gens.Create(context.Background(), user, CreateInput{
		ModelID: "fal/kling-3.0",
		Prompt:  "anything",
	})
	if err == nil {
		t.Fatal("created a generation on a locked model")
	}

	var after models.User
	db.First(&after, "id = ?", user.ID)
	if after.Credits != 100 {
		t.Fatalf("credits moved on a rejected job: %d", after.Credits)
	}
}

func TestCreateRejectsInsufficientCredits(t *testing.T) {
	db := testDB(t)
	user := newUser(t, db, 0)
	gens, _ := harness(t, db)

	if _, err := gens.Create(context.Background(), user, CreateInput{
		ModelID: "mock/motion", // costs 5
		Prompt:  "anything",
	}); err == nil {
		t.Fatal("created a generation with no credits")
	}
}

func TestCreateRejectsEmptyPrompt(t *testing.T) {
	db := testDB(t)
	user := newUser(t, db, 100)
	gens, _ := harness(t, db)

	if _, err := gens.Create(context.Background(), user, CreateInput{
		ModelID: "mock/still",
		Prompt:  "   ",
	}); err == nil {
		t.Fatal("accepted a blank prompt")
	}
}

// Signing in twice must land on the same row.
//
// Regression test. The upsert generates a fresh UUID on every call; on a
// conflict Postgres keeps the existing row's id, so without a RETURNING clause
// the struct came back holding an id that was never written. Every request
// after the first login then referenced a user that did not exist: BYOK lookups
// found nothing, and creating a generation failed on the users foreign key.
func TestUpsertReturnsThePersistedUser(t *testing.T) {
	db := testDB(t)
	users := NewUsers(db)
	ctx := context.Background()

	claims := jwtx.Claims{
		Subject: "auth0|stable-" + time.Now().Format("150405.000000000"),
		Email:   "first@example.com",
		Name:    "First Name",
	}

	first, err := users.Upsert(ctx, claims)
	if err != nil {
		t.Fatalf("first upsert: %v", err)
	}
	t.Cleanup(func() { db.Unscoped().Delete(&models.User{}, "id = ?", first.ID) })

	claims.Email = "second@example.com"
	claims.Name = "Second Name"
	second, err := users.Upsert(ctx, claims)
	if err != nil {
		t.Fatalf("second upsert: %v", err)
	}

	if first.ID != second.ID {
		t.Fatalf("id changed across sign-ins: %s then %s", first.ID, second.ID)
	}

	// The returned id must actually exist, which is what the foreign key needs.
	var count int64
	db.Model(&models.User{}).Where("id = ?", second.ID).Count(&count)
	if count != 1 {
		t.Fatalf("upsert returned an id with no row behind it: %s", second.ID)
	}

	// And exactly one row, not a duplicate per sign-in.
	db.Model(&models.User{}).Where("auth0_subject = ?", claims.Subject).Count(&count)
	if count != 1 {
		t.Fatalf("expected 1 row for the subject, found %d", count)
	}

	if second.Email != "second@example.com" || second.Name != "Second Name" {
		t.Fatalf("profile not refreshed from the token: %+v", second)
	}

	// The end that actually broke in production: a row referencing this id.
	gen := models.Generation{
		UserID: second.ID, ProviderID: "mock", ModelID: "mock/still",
		Modality: "image", Prompt: "fk check", Status: models.StatusQueued,
	}
	if err := db.Create(&gen).Error; err != nil {
		t.Fatalf("foreign key rejected the returned user id: %v", err)
	}
	db.Unscoped().Delete(&gen)
}

// Credits must survive a re-login, or signing out and in again would be free
// top-ups.
func TestUpsertDoesNotResetCredits(t *testing.T) {
	db := testDB(t)
	users := NewUsers(db)
	ctx := context.Background()
	claims := jwtx.Claims{Subject: "auth0|credits-" + time.Now().Format("150405.000000000")}

	u, err := users.Upsert(ctx, claims)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Unscoped().Delete(&models.User{}, "id = ?", u.ID) })

	if err := db.Model(u).UpdateColumn("credits", 7).Error; err != nil {
		t.Fatal(err)
	}

	again, err := users.Upsert(ctx, claims)
	if err != nil {
		t.Fatal(err)
	}
	if again.Credits != 7 {
		t.Fatalf("re-login changed the balance: want 7, got %d", again.Credits)
	}
}
