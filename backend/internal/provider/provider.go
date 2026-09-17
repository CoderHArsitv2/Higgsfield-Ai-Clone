// Package provider is the seam between this app and every upstream model
// vendor. Adding a vendor means adding one file that implements Provider and
// registering it -- nothing else in the codebase changes.
package provider

import (
	"context"
	"errors"
)

type Modality string

const (
	ModalityImage Modality = "image"
	ModalityVideo Modality = "video"
	ModalityAudio Modality = "audio"
)

// ParamKind drives how the frontend renders a control, so a model can expose
// its own options without the UI hardcoding anything per model.
type ParamKind string

const (
	ParamSelect ParamKind = "select"
	ParamNumber ParamKind = "number"
	ParamText   ParamKind = "text"
	ParamBool   ParamKind = "bool"
)

type Option struct {
	Value string `json:"value"`
	Label string `json:"label"`
}

type ParamSpec struct {
	Key      string    `json:"key"`
	Label    string    `json:"label"`
	Kind     ParamKind `json:"kind"`
	Options  []Option  `json:"options,omitempty"`
	Default  any       `json:"default,omitempty"`
	Min      *float64  `json:"min,omitempty"`
	Max      *float64  `json:"max,omitempty"`
	Step     *float64  `json:"step,omitempty"`
	Optional bool      `json:"optional,omitempty"`
	Help     string    `json:"help,omitempty"`
}

type ModelSpec struct {
	ID          string      `json:"id"`
	ProviderID  string      `json:"provider_id"`
	Name        string      `json:"name"`
	Modality    Modality    `json:"modality"`
	Description string      `json:"description"`
	Tags        []string    `json:"tags,omitempty"`
	CreditCost  int         `json:"credit_cost"`
	Params      []ParamSpec `json:"params,omitempty"`
	RefImages   int         `json:"ref_images"`
	Featured    bool        `json:"featured,omitempty"`
}

type SubmitRequest struct {
	Model     ModelSpec
	Prompt    string
	Params    map[string]any
	RefImages []string
	APIKey    string // already resolved: the user's own key, or the server's
}

type SubmitResult struct {
	ExternalID string
	Done       bool
	Assets     []ResultAsset
}

type PollRequest struct {
	Model      ModelSpec
	ExternalID string
	APIKey     string
}

type PollResult struct {
	Done     bool
	Failed   bool
	Error    string
	Progress int
	Assets   []ResultAsset
}

type ResultAsset struct {
	Kind         string
	URL          string
	ThumbnailURL string
	Width        int
	Height       int
	DurationMS   int
}

type Provider interface {
	ID() string
	Name() string
	// EnvKey is the environment variable holding the server-wide credential.
	// A provider whose EnvKey is unset is locked unless the user brings a key.
	EnvKey() string
	DocsURL() string
	Models() []ModelSpec
	Submit(ctx context.Context, req SubmitRequest) (SubmitResult, error)
	Poll(ctx context.Context, req PollRequest) (PollResult, error)
}

var (
	ErrUnsupportedModel = errors.New("model is not supported by this provider")
	ErrMissingKey       = errors.New("no api key available for this provider")
)

// helpers for declaring specs concisely
func f64(v float64) *float64 { return &v }

func opts(pairs ...string) []Option {
	out := make([]Option, 0, len(pairs)/2)
	for i := 0; i+1 < len(pairs); i += 2 {
		out = append(out, Option{Value: pairs[i], Label: pairs[i+1]})
	}
	return out
}

// aspectParam is shared by nearly every visual model.
func aspectParam(def string, values ...string) ParamSpec {
	if len(values) == 0 {
		values = []string{"16:9", "Landscape 16:9", "9:16", "Portrait 9:16", "1:1", "Square 1:1", "21:9", "Cinematic 21:9", "4:3", "Classic 4:3"}
	}
	return ParamSpec{Key: "aspect_ratio", Label: "Aspect ratio", Kind: ParamSelect, Options: opts(values...), Default: def}
}

func durationParam(def float64, options ...string) ParamSpec {
	if len(options) == 0 {
		options = []string{"5", "5 seconds", "10", "10 seconds"}
	}
	return ParamSpec{Key: "duration", Label: "Duration", Kind: ParamSelect, Options: opts(options...), Default: def}
}

func resolutionParam(def string, values ...string) ParamSpec {
	if len(values) == 0 {
		values = []string{"720p", "720p", "1080p", "1080p"}
	}
	return ParamSpec{Key: "resolution", Label: "Resolution", Kind: ParamSelect, Options: opts(values...), Default: def}
}

var negativePromptParam = ParamSpec{
	Key: "negative_prompt", Label: "Negative prompt", Kind: ParamText, Optional: true,
	Help: "What to keep out of the result",
}

var seedParam = ParamSpec{
	Key: "seed", Label: "Seed", Kind: ParamNumber, Optional: true, Min: f64(0), Max: f64(2147483647),
	Help: "Reuse a seed to reproduce a result",
}
