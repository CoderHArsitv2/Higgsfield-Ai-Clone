package provider

import (
	"context"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/coderHArsitv2/higgsfield-clone/backend/internal/storage"
	"github.com/coderHArsitv2/higgsfield-clone/backend/pkg/httpx"
)

// Google covers Gemini image models and Veo video. Gemini authenticates with a
// query parameter rather than a header, which is why it does not reuse the
// bearer-token helper.
type Google struct{ store storage.Storage }

func NewGoogle(store storage.Storage) *Google { return &Google{store: store} }

func (g *Google) ID() string      { return "google" }
func (g *Google) Name() string    { return "Google AI" }
func (g *Google) EnvKey() string  { return "GEMINI_API_KEY" }
func (g *Google) DocsURL() string { return "https://aistudio.google.com/apikey" }

var googleUpstream = map[string]string{
	"google/nano-banana-pro": "gemini-3-pro-image-preview",
	"google/nano-banana":     "gemini-2.5-flash-image",
	"google/veo-3.1":         "veo-3.1-generate-preview",
}

func (g *Google) Models() []ModelSpec {
	return []ModelSpec{
		{ID: "google/nano-banana-pro", Name: "Nano Banana Pro", Modality: ModalityImage, CreditCost: 6, Featured: true,
			Description: "Gemini 3 Pro Image. Exceptional at targeted edits that leave the rest of the frame alone.",
			Tags:        []string{"image", "editing"},
			Params:      []ParamSpec{aspectParam("1:1"), negativePromptParam}, RefImages: 3},
		{ID: "google/nano-banana", Name: "Nano Banana", Modality: ModalityImage, CreditCost: 2,
			Description: "Fast Gemini image model for iteration.", Tags: []string{"image", "fast"},
			Params: []ParamSpec{aspectParam("1:1")}, RefImages: 3},
		{ID: "google/veo-3.1", Name: "Veo 3.1", Modality: ModalityVideo, CreditCost: 30, Featured: true,
			Description: "4K cinematic generation with native synced audio.", Tags: []string{"video", "audio"},
			Params:    []ParamSpec{aspectParam("16:9", "16:9", "Landscape 16:9", "9:16", "Portrait 9:16"), negativePromptParam},
			RefImages: 1},
	}
}

func (g *Google) client() *httpx.Client {
	return httpx.New("https://generativelanguage.googleapis.com", 180*time.Second)
}

func (g *Google) Submit(ctx context.Context, req SubmitRequest) (SubmitResult, error) {
	upstream, ok := googleUpstream[req.Model.ID]
	if !ok {
		return SubmitResult{}, ErrUnsupportedModel
	}
	if req.Model.Modality == ModalityVideo {
		return g.submitVeo(ctx, req, upstream)
	}
	return g.submitImage(ctx, req, upstream)
}

func (g *Google) submitImage(ctx context.Context, req SubmitRequest, upstream string) (SubmitResult, error) {
	body := map[string]any{
		"contents": []map[string]any{{
			"parts": []map[string]any{{"text": req.Prompt}},
		}},
	}
	var res struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					InlineData *struct {
						MimeType string `json:"mimeType"`
						Data     string `json:"data"`
					} `json:"inlineData"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}
	path := "/v1beta/models/" + upstream + ":generateContent?key=" + req.APIKey
	if err := g.client().JSON(ctx, "POST", path, body, &res); err != nil {
		return SubmitResult{}, err
	}

	var assets []ResultAsset
	for ci, cand := range res.Candidates {
		for pi, part := range cand.Content.Parts {
			if part.InlineData == nil || part.InlineData.Data == "" {
				continue
			}
			raw, err := base64.StdEncoding.DecodeString(part.InlineData.Data)
			if err != nil {
				continue
			}
			url, err := g.store.Put(ctx, fmt.Sprintf("gemini-%d-%d.png", ci, pi), raw, part.InlineData.MimeType)
			if err != nil {
				return SubmitResult{}, err
			}
			assets = append(assets, ResultAsset{Kind: "image", URL: url})
		}
	}
	if len(assets) == 0 {
		return SubmitResult{}, fmt.Errorf("gemini returned no image data")
	}
	return SubmitResult{Done: true, Assets: assets}, nil
}

func (g *Google) submitVeo(ctx context.Context, req SubmitRequest, upstream string) (SubmitResult, error) {
	params := map[string]any{}
	for k, v := range req.Params {
		params[k] = v
	}
	body := map[string]any{
		"instances":  []map[string]any{{"prompt": req.Prompt}},
		"parameters": params,
	}
	var res struct {
		Name string `json:"name"`
	}
	path := "/v1beta/models/" + upstream + ":predictLongRunning?key=" + req.APIKey
	if err := g.client().JSON(ctx, "POST", path, body, &res); err != nil {
		return SubmitResult{}, err
	}
	if res.Name == "" {
		return SubmitResult{}, fmt.Errorf("veo did not return an operation name")
	}
	return SubmitResult{ExternalID: res.Name}, nil
}

func (g *Google) Poll(ctx context.Context, req PollRequest) (PollResult, error) {
	var res struct {
		Done  bool `json:"done"`
		Error *struct {
			Message string `json:"message"`
		} `json:"error"`
		Response struct {
			GenerateVideoResponse struct {
				GeneratedSamples []struct {
					Video struct {
						URI string `json:"uri"`
					} `json:"video"`
				} `json:"generatedSamples"`
			} `json:"generateVideoResponse"`
		} `json:"response"`
	}
	path := "/v1beta/" + req.ExternalID + "?key=" + req.APIKey
	if err := g.client().JSON(ctx, "GET", path, nil, &res); err != nil {
		return PollResult{}, err
	}
	if res.Error != nil && res.Error.Message != "" {
		return PollResult{Failed: true, Error: res.Error.Message}, nil
	}
	if !res.Done {
		return PollResult{Progress: 40}, nil
	}

	var assets []ResultAsset
	for _, s := range res.Response.GenerateVideoResponse.GeneratedSamples {
		if s.Video.URI != "" {
			// Veo file URIs need the key appended to be fetchable.
			assets = append(assets, ResultAsset{Kind: "video", URL: s.Video.URI + "&key=" + req.APIKey})
		}
	}
	if len(assets) == 0 {
		return PollResult{Failed: true, Error: "veo finished without returning a video"}, nil
	}
	return PollResult{Done: true, Progress: 100, Assets: assets}, nil
}
