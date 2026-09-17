package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/coderHArsitv2/higgsfield-clone/backend/internal/storage"
	"github.com/coderHArsitv2/higgsfield-clone/backend/pkg/httpx"
)

// ElevenLabs covers the voice pillar. Synthesis is synchronous, so these jobs
// complete inside Submit and never queue.
type ElevenLabs struct{ store storage.Storage }

func NewElevenLabs(store storage.Storage) *ElevenLabs { return &ElevenLabs{store: store} }

func (e *ElevenLabs) ID() string      { return "elevenlabs" }
func (e *ElevenLabs) Name() string    { return "ElevenLabs" }
func (e *ElevenLabs) EnvKey() string  { return "ELEVENLABS_API_KEY" }
func (e *ElevenLabs) DocsURL() string { return "https://elevenlabs.io/app/settings/api-keys" }

// Stock voice ids published by ElevenLabs.
var elevenVoices = []Option{
	{Value: "21m00Tcm4TlvDq8ikWAM", Label: "Rachel — calm narration"},
	{Value: "AZnzlk1XvdvUeBnXmlld", Label: "Domi — confident"},
	{Value: "EXAVITQu4vr4xnSDxMaL", Label: "Sarah — soft"},
	{Value: "onwK4e9ZLuTAKqWW03F9", Label: "Daniel — broadcast"},
	{Value: "pNInz6obpgDQGcFmaJgB", Label: "Adam — deep"},
}

func (e *ElevenLabs) Models() []ModelSpec {
	params := []ParamSpec{
		{Key: "voice_id", Label: "Voice", Kind: ParamSelect, Default: elevenVoices[0].Value, Options: elevenVoices},
		{Key: "stability", Label: "Stability", Kind: ParamNumber, Default: 0.5, Min: f64(0), Max: f64(1), Step: f64(0.05)},
		{Key: "similarity_boost", Label: "Similarity", Kind: ParamNumber, Default: 0.75, Min: f64(0), Max: f64(1), Step: f64(0.05)},
	}
	return []ModelSpec{
		{ID: "elevenlabs/multilingual-v2", Name: "Eleven Multilingual v2", Modality: ModalityAudio, CreditCost: 2, Featured: true,
			Description: "Highest quality speech across 29 languages.", Tags: []string{"audio", "voice"}, Params: params},
		{ID: "elevenlabs/turbo-v2-5", Name: "Eleven Turbo v2.5", Modality: ModalityAudio, CreditCost: 1,
			Description: "Low latency speech for drafts and long scripts.", Tags: []string{"audio", "fast"}, Params: params},
	}
}

var elevenUpstream = map[string]string{
	"elevenlabs/multilingual-v2": "eleven_multilingual_v2",
	"elevenlabs/turbo-v2-5":      "eleven_turbo_v2_5",
}

func (e *ElevenLabs) Submit(ctx context.Context, req SubmitRequest) (SubmitResult, error) {
	modelID, ok := elevenUpstream[req.Model.ID]
	if !ok {
		return SubmitResult{}, ErrUnsupportedModel
	}
	voice, _ := req.Params["voice_id"].(string)
	if voice == "" {
		voice = elevenVoices[0].Value
	}

	settings := map[string]any{}
	for _, k := range []string{"stability", "similarity_boost"} {
		if v, ok := req.Params[k]; ok {
			settings[k] = v
		}
	}
	body := map[string]any{"text": req.Prompt, "model_id": modelID}
	if len(settings) > 0 {
		body["voice_settings"] = settings
	}

	c := httpx.New("https://api.elevenlabs.io", 120*time.Second).WithHeader("xi-api-key", req.APIKey)
	raw, ctype, err := c.Bytes(ctx, "POST", "/v1/text-to-speech/"+voice, body)
	if err != nil {
		return SubmitResult{}, err
	}
	if len(raw) == 0 {
		return SubmitResult{}, fmt.Errorf("elevenlabs returned empty audio")
	}
	if ctype == "" {
		ctype = "audio/mpeg"
	}
	url, err := e.store.Put(ctx, "speech.mp3", raw, ctype)
	if err != nil {
		return SubmitResult{}, err
	}
	return SubmitResult{Done: true, Assets: []ResultAsset{{Kind: "audio", URL: url}}}, nil
}

func (e *ElevenLabs) Poll(context.Context, PollRequest) (PollResult, error) {
	// Synthesis finishes inside Submit, so a poll can only mean a lost job.
	return PollResult{Failed: true, Error: "elevenlabs jobs complete synchronously"}, nil
}
