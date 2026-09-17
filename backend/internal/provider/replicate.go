package provider

import (
	"context"
	"time"

	"github.com/coderHArsitv2/higgsfield-clone/backend/pkg/httpx"
)

// Replicate is the second aggregator. It overlaps fal deliberately: if one
// vendor is down or rate limited, the same model is reachable through the other.
type Replicate struct{}

func NewReplicate() *Replicate { return &Replicate{} }

func (r *Replicate) ID() string      { return "replicate" }
func (r *Replicate) Name() string    { return "Replicate" }
func (r *Replicate) EnvKey() string  { return "REPLICATE_API_TOKEN" }
func (r *Replicate) DocsURL() string { return "https://replicate.com/account/api-tokens" }

var replicateEndpoints = map[string]string{
	"replicate/flux-1.1-pro":  "black-forest-labs/flux-1.1-pro",
	"replicate/flux-schnell":  "black-forest-labs/flux-schnell",
	"replicate/sdxl":          "stability-ai/sdxl",
	"replicate/kling-v2":      "kwaivgi/kling-v2.1",
	"replicate/hunyuan-video": "tencent/hunyuan-video",
	"replicate/ltx-video":     "lightricks/ltx-video",
	"replicate/minimax-video": "minimax/video-01",
	"replicate/musicgen":      "meta/musicgen",
}

func (r *Replicate) Models() []ModelSpec {
	return []ModelSpec{
		{ID: "replicate/flux-1.1-pro", Name: "FLUX 1.1 Pro", Modality: ModalityImage, CreditCost: 5,
			Description: "Flagship FLUX via Replicate.", Tags: []string{"image"},
			Params: []ParamSpec{aspectParam("1:1"), negativePromptParam, seedParam}, RefImages: 1},
		{ID: "replicate/flux-schnell", Name: "FLUX schnell", Modality: ModalityImage, CreditCost: 1,
			Description: "Four-step model. Near-instant drafts.", Tags: []string{"image", "fast"},
			Params: []ParamSpec{aspectParam("1:1"), seedParam}},
		{ID: "replicate/sdxl", Name: "SDXL", Modality: ModalityImage, CreditCost: 2,
			Description: "Stable Diffusion XL with a large LoRA ecosystem.", Tags: []string{"image"},
			Params: []ParamSpec{aspectParam("1:1"), negativePromptParam, seedParam}},
		{ID: "replicate/kling-v2", Name: "Kling v2.1", Modality: ModalityVideo, CreditCost: 18,
			Description: "Kling via Replicate.", Tags: []string{"video"},
			Params: []ParamSpec{aspectParam("16:9"), durationParam(5), negativePromptParam}, RefImages: 1},
		{ID: "replicate/hunyuan-video", Name: "Hunyuan Video", Modality: ModalityVideo, CreditCost: 15,
			Description: "Open-weights video model with strong motion.", Tags: []string{"video", "open"},
			Params: []ParamSpec{aspectParam("16:9"), durationParam(5)}},
		{ID: "replicate/ltx-video", Name: "LTX Video", Modality: ModalityVideo, CreditCost: 8,
			Description: "Real-time class video generation.", Tags: []string{"video", "fast"},
			Params: []ParamSpec{aspectParam("16:9"), durationParam(5)}},
		{ID: "replicate/minimax-video", Name: "MiniMax Video 01", Modality: ModalityVideo, CreditCost: 16,
			Description: "Cinematic motion with strong prompt adherence.", Tags: []string{"video"},
			Params: []ParamSpec{aspectParam("16:9")}, RefImages: 1},
		{ID: "replicate/musicgen", Name: "MusicGen", Modality: ModalityAudio, CreditCost: 3,
			Description: "Text to music for background beds.", Tags: []string{"audio"},
			Params: []ParamSpec{{Key: "duration", Label: "Duration (s)", Kind: ParamNumber, Default: 8, Min: f64(2), Max: f64(30)}}},
	}
}

func (r *Replicate) client(key string) *httpx.Client {
	return httpx.New("https://api.replicate.com", 60*time.Second).
		WithHeader("Authorization", "Bearer "+key).
		WithHeader("Prefer", "wait=0")
}

func (r *Replicate) Submit(ctx context.Context, req SubmitRequest) (SubmitResult, error) {
	slug, ok := replicateEndpoints[req.Model.ID]
	if !ok {
		return SubmitResult{}, ErrUnsupportedModel
	}
	input := map[string]any{"prompt": req.Prompt}
	for k, v := range req.Params {
		input[k] = v
	}
	if len(req.RefImages) > 0 {
		input["image"] = req.RefImages[0]
	}

	var res struct {
		ID string `json:"id"`
	}
	if err := r.client(req.APIKey).JSON(ctx, "POST", "/v1/models/"+slug+"/predictions",
		map[string]any{"input": input}, &res); err != nil {
		return SubmitResult{}, err
	}
	return SubmitResult{ExternalID: res.ID}, nil
}

func (r *Replicate) Poll(ctx context.Context, req PollRequest) (PollResult, error) {
	var res struct {
		Status string `json:"status"`
		Error  string `json:"error"`
		Output any    `json:"output"`
	}
	if err := r.client(req.APIKey).JSON(ctx, "GET", "/v1/predictions/"+req.ExternalID, nil, &res); err != nil {
		return PollResult{}, err
	}

	switch res.Status {
	case "starting":
		return PollResult{Progress: 5}, nil
	case "processing":
		return PollResult{Progress: 50}, nil
	case "failed", "canceled":
		msg := res.Error
		if msg == "" {
			msg = "replicate reported " + res.Status
		}
		return PollResult{Failed: true, Error: msg}, nil
	}

	kind := string(req.Model.Modality)
	var assets []ResultAsset
	for _, u := range flattenURLs(res.Output) {
		assets = append(assets, ResultAsset{Kind: kind, URL: u})
	}
	if len(assets) == 0 {
		return PollResult{Failed: true, Error: "replicate returned no output"}, nil
	}
	return PollResult{Done: true, Progress: 100, Assets: assets}, nil
}

// Replicate output is a string, a list of strings, or an object with a url --
// depending on the model. Normalise all three.
func flattenURLs(v any) []string {
	switch t := v.(type) {
	case string:
		if t != "" {
			return []string{t}
		}
	case []any:
		var out []string
		for _, item := range t {
			out = append(out, flattenURLs(item)...)
		}
		return out
	case map[string]any:
		if u, ok := t["url"].(string); ok && u != "" {
			return []string{u}
		}
	}
	return nil
}
