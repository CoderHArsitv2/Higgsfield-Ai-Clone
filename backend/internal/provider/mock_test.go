package provider

import (
	"strings"
	"testing"
)

// The sandbox is what a visitor sees before they own a single API key, and this
// is a video product. Every sandbox model except the voice one must therefore
// hand back a playable clip -- a still would misrepresent what the app does.
func TestSandboxReturnsVideoForEveryVisualModel(t *testing.T) {
	m := NewMock()

	for _, spec := range m.Models() {
		if spec.Modality == ModalityAudio {
			continue
		}

		assets := m.assets(spec, hash(spec.ID))
		if len(assets) != 1 {
			t.Fatalf("%s: expected one asset, got %d", spec.ID, len(assets))
		}

		a := assets[0]
		if a.Kind != "video" {
			t.Errorf("%s: asset kind = %q, want \"video\"", spec.ID, a.Kind)
		}
		if !strings.HasSuffix(a.URL, ".mp4") {
			t.Errorf("%s: url %q is not a clip", spec.ID, a.URL)
		}
		// Site-relative, so the browser resolves it against the app's own
		// origin and the sandbox depends on nothing external.
		if !strings.HasPrefix(a.URL, "/showcase/") {
			t.Errorf("%s: url %q is not served from the app itself", spec.ID, a.URL)
		}
		if a.ThumbnailURL == "" {
			t.Errorf("%s: no poster, so the card is blank until the clip decodes", spec.ID)
		}
	}
}

// Audio is the one modality a clip cannot stand in for.
func TestSandboxVoiceStaysAudio(t *testing.T) {
	m := NewMock()
	for _, spec := range m.Models() {
		if spec.Modality != ModalityAudio {
			continue
		}
		if got := m.assets(spec, "seed")[0].Kind; got != "audio" {
			t.Errorf("%s: asset kind = %q, want \"audio\"", spec.ID, got)
		}
	}
}

// The same prompt must keep returning the same clip: a gallery that reshuffles
// its own history on every poll looks broken.
func TestSandboxAssetsAreStableForASeed(t *testing.T) {
	m := NewMock()
	spec := m.Models()[0]
	first := m.assets(spec, "abc123")[0].URL
	for i := 0; i < 5; i++ {
		if got := m.assets(spec, "abc123")[0].URL; got != first {
			t.Fatalf("same seed returned %q then %q", first, got)
		}
	}
}
