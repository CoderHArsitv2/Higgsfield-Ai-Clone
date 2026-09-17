package provider

import "testing"

func catalogFor(t *testing.T, env map[string]string, userKeys map[string]bool) map[string]ModelView {
	t.Helper()
	r := NewRegistry(env, NewMock(), NewFal(), NewOpenAI(nil))
	out := map[string]ModelView{}
	for _, m := range r.Catalog(userKeys) {
		out[m.ID] = m
	}
	return out
}

// The sandbox provider has no env key, so it must never be locked -- that is
// what lets a brand new install be used before any credential exists.
func TestSandboxIsAlwaysUnlocked(t *testing.T) {
	c := catalogFor(t, map[string]string{}, nil)
	if !c["mock/still"].Enabled {
		t.Fatal("sandbox model was locked")
	}
	if c["mock/still"].Access != AccessServer {
		t.Fatalf("expected server access, got %q", c["mock/still"].Access)
	}
}

func TestModelsLockWithoutAnyKey(t *testing.T) {
	c := catalogFor(t, map[string]string{}, nil)
	m := c["fal/kling-3.0"]
	if m.Enabled {
		t.Fatal("fal model was enabled with no key anywhere")
	}
	if m.Access != AccessLocked {
		t.Fatalf("expected locked, got %q", m.Access)
	}
	if m.LockReason == "" {
		t.Fatal("locked model gave the user no reason")
	}
}

func TestServerKeyUnlocksEveryModelOfThatProvider(t *testing.T) {
	c := catalogFor(t, map[string]string{"FAL_KEY": "k"}, nil)
	if !c["fal/kling-3.0"].Enabled || c["fal/kling-3.0"].Access != AccessServer {
		t.Fatal("server key did not unlock the provider")
	}
	// A key for one provider must not unlock another.
	if c["openai/sora-2"].Enabled {
		t.Fatal("a fal key unlocked an openai model")
	}
}

func TestUserKeyUnlocksWithoutServerKey(t *testing.T) {
	c := catalogFor(t, map[string]string{}, map[string]bool{"openai": true})
	m := c["openai/sora-2"]
	if !m.Enabled {
		t.Fatal("BYOK did not unlock the model")
	}
	if m.Access != AccessBYOK {
		t.Fatalf("expected byok access, got %q", m.Access)
	}
}

// A user's own key must win over the platform's, so BYOK also works as an
// escape hatch from shared rate limits.
func TestResolveKeyPrefersUserKey(t *testing.T) {
	r := NewRegistry(map[string]string{"FAL_KEY": "server-key"}, NewFal())

	key, own, err := r.ResolveKey("fal", "user-key")
	if err != nil || key != "user-key" || !own {
		t.Fatalf("got %q own=%v err=%v", key, own, err)
	}

	key, own, err = r.ResolveKey("fal", "")
	if err != nil || key != "server-key" || own {
		t.Fatalf("got %q own=%v err=%v", key, own, err)
	}

	if _, _, err := r.ResolveKey("openai", ""); err == nil {
		t.Fatal("resolved a key for a provider with no credential")
	}
}

func TestCatalogSortsEnabledFirst(t *testing.T) {
	r := NewRegistry(map[string]string{}, NewMock(), NewFal())
	seenLocked := false
	for _, m := range r.Catalog(nil) {
		if !m.Enabled {
			seenLocked = true
			continue
		}
		if seenLocked {
			t.Fatal("an enabled model was sorted after a locked one")
		}
	}
}

// The sandbox needs no credential, so resolving one must succeed and must not
// be reported as the user's own key (that would skip credit accounting).
func TestResolveKeyForKeylessProvider(t *testing.T) {
	r := NewRegistry(map[string]string{}, NewMock())
	key, own, err := r.ResolveKey("mock", "")
	if err != nil {
		t.Fatalf("keyless provider failed to resolve: %v", err)
	}
	if key != "" || own {
		t.Fatalf("got key=%q own=%v, want empty and not-own", key, own)
	}
}
