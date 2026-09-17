package provider

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/coderHArsitv2/higgsfield-clone/backend/pkg/httpx"
)

// Fal routes to fal.ai's queue API. One credential here unlocks most of the
// catalogue, which is why it is the key worth adding first.
type Fal struct{}

func NewFal() *Fal { return &Fal{} }

func (f *Fal) ID() string      { return "fal" }
func (f *Fal) Name() string    { return "fal.ai" }
func (f *Fal) EnvKey() string  { return "FAL_KEY" }
func (f *Fal) DocsURL() string { return "https://fal.ai/dashboard/keys" }

// endpoint is the upstream route for a model id in our catalogue.
var falEndpoints = map[string]string{
	"fal/kling-3.0":       "fal-ai/kling-video/v3/standard/text-to-video",
	"fal/kling-2.6":       "fal-ai/kling-video/v2.6/standard/text-to-video",
	"fal/seedance-2.0":    "fal-ai/bytedance/seedance/v2/text-to-video",
	"fal/wan-2.7":         "fal-ai/wan/v2.7/text-to-video",
	"fal/veo-3.1":         "fal-ai/veo3.1",
	"fal/sora-2":          "fal-ai/sora-2/text-to-video",
	"fal/minimax-hailuo":  "fal-ai/minimax/hailuo-02/standard/text-to-video",
	"fal/luma-ray-3":      "fal-ai/luma-dream-machine/ray-3",
	"fal/flux-pro-1.1":    "fal-ai/flux-pro/v1.1-ultra",
	"fal/flux-dev":        "fal-ai/flux/dev",
	"fal/nano-banana-pro": "fal-ai/gemini-3-pro-image",
	"fal/recraft-v3":      "fal-ai/recraft-v3",
	"fal/ideogram-v3":     "fal-ai/ideogram/v3",
	"fal/qwen-image":      "fal-ai/qwen-image",
}

func (f *Fal) Models() []ModelSpec {
	video := func(id, name, desc string, cost int, featured bool, params ...ParamSpec) ModelSpec {
		if len(params) == 0 {
			params = []ParamSpec{aspectParam("16:9"), durationParam(5), resolutionParam("1080p"), negativePromptParam, seedParam}
		}
		return ModelSpec{ID: id, Name: name, Modality: ModalityVideo, Description: desc, CreditCost: cost,
			Featured: featured, Tags: []string{"video"}, Params: params, RefImages: 1}
	}
	image := func(id, name, desc string, cost int, featured bool) ModelSpec {
		return ModelSpec{ID: id, Name: name, Modality: ModalityImage, Description: desc, CreditCost: cost,
			Featured: featured, Tags: []string{"image"},
			Params:    []ParamSpec{aspectParam("1:1"), {Key: "num_images", Label: "Images", Kind: ParamNumber, Default: 1, Min: f64(1), Max: f64(4)}, negativePromptParam, seedParam},
			RefImages: 3}
	}

	return []ModelSpec{
		video("fal/kling-3.0", "Kling 3.0", "Photorealism with complex motion. The current standard for character work.", 20, true),
		video("fal/seedance-2.0", "Seedance 2.0", "Native audio-video: synced lip-sync, SFX and music in one pass.", 24, true),
		video("fal/veo-3.1", "Veo 3.1", "Crystal clear 4K with native cinematic flow and audio.", 30, true),
		video("fal/sora-2", "Sora 2", "Deep world simulation with accurate physics and object permanence.", 28, false),
		video("fal/wan-2.7", "Wan 2.7", "The speed/richness balance point. Good default for iteration.", 12, false),
		video("fal/kling-2.6", "Kling 2.6", "Proven engine for fast, stable character animation.", 14, false),
		video("fal/minimax-hailuo", "Hailuo 02", "Strong prompt adherence on stylised motion.", 14, false),
		video("fal/luma-ray-3", "Ray 3", "Fluid natural motion with strong scene coherence.", 18, false),
		image("fal/nano-banana-pro", "Nano Banana Pro", "Gemini 3 Pro image. Best-in-class instruction following on edits.", 6, true),
		image("fal/flux-pro-1.1", "FLUX 1.1 Pro Ultra", "High fidelity stills at up to 4MP.", 5, true),
		image("fal/flux-dev", "FLUX dev", "Fast, cheap iteration model.", 2, false),
		image("fal/recraft-v3", "Recraft V3", "Vector-aware generation with reliable in-image text.", 4, false),
		image("fal/ideogram-v3", "Ideogram V3", "Typography and poster layouts.", 4, false),
		image("fal/qwen-image", "Qwen Image", "Strong bilingual text rendering.", 3, false),
	}
}

func (f *Fal) client(key string) *httpx.Client {
	return httpx.New("https://queue.fal.run", 60*time.Second).
		WithHeader("Authorization", "Key "+key)
}

func (f *Fal) Submit(ctx context.Context, req SubmitRequest) (SubmitResult, error) {
	endpoint, ok := falEndpoints[req.Model.ID]
	if !ok {
		return SubmitResult{}, ErrUnsupportedModel
	}

	body := map[string]any{"prompt": req.Prompt}
	for k, v := range req.Params {
		body[k] = v
	}
	if len(req.RefImages) > 0 {
		body["image_url"] = req.RefImages[0]
	}

	var res struct {
		RequestID string `json:"request_id"`
	}
	if err := f.client(req.APIKey).JSON(ctx, "POST", "/"+endpoint, body, &res); err != nil {
		return SubmitResult{}, err
	}
	if res.RequestID == "" {
		return SubmitResult{}, fmt.Errorf("fal did not return a request id")
	}
	// The status route needs the endpoint too, so carry both in the handle.
	return SubmitResult{ExternalID: endpoint + "|" + res.RequestID}, nil
}

func (f *Fal) Poll(ctx context.Context, req PollRequest) (PollResult, error) {
	endpoint, id, ok := strings.Cut(req.ExternalID, "|")
	if !ok {
		return PollResult{Failed: true, Error: "malformed fal job handle"}, nil
	}
	c := f.client(req.APIKey)

	var status struct {
		Status string `json:"status"`
	}
	if err := c.JSON(ctx, "GET", "/"+endpoint+"/requests/"+id+"/status", nil, &status); err != nil {
		return PollResult{}, err
	}

	switch status.Status {
	case "IN_QUEUE":
		return PollResult{Progress: 5}, nil
	case "IN_PROGRESS":
		return PollResult{Progress: 50}, nil
	case "COMPLETED":
		// fall through
	default:
		return PollResult{Failed: true, Error: "fal reported status " + status.Status}, nil
	}

	var out falResult
	if err := c.JSON(ctx, "GET", "/"+endpoint+"/requests/"+id, nil, &out); err != nil {
		return PollResult{}, err
	}
	assets := out.assets()
	if len(assets) == 0 {
		return PollResult{Failed: true, Error: "fal returned no output"}, nil
	}
	return PollResult{Done: true, Progress: 100, Assets: assets}, nil
}

type falFile struct {
	URL    string `json:"url"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

type falResult struct {
	Images []falFile `json:"images"`
	Video  *falFile  `json:"video"`
	Audio  *falFile  `json:"audio"`
}

func (r falResult) assets() []ResultAsset {
	var out []ResultAsset
	if r.Video != nil && r.Video.URL != "" {
		out = append(out, ResultAsset{Kind: "video", URL: r.Video.URL, Width: r.Video.Width, Height: r.Video.Height})
	}
	if r.Audio != nil && r.Audio.URL != "" {
		out = append(out, ResultAsset{Kind: "audio", URL: r.Audio.URL})
	}
	for _, im := range r.Images {
		if im.URL != "" {
			out = append(out, ResultAsset{Kind: "image", URL: im.URL, Width: im.Width, Height: im.Height})
		}
	}
	return out
}
