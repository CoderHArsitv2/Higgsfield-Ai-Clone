package provider

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Mock is a fully working provider that needs no credentials. It exists so the
// whole product -- auth, job queue, polling, gallery, credits -- can be
// exercised end to end before a single real API key is added, and so the
// polling worker is tested against something that genuinely takes time to
// finish rather than returning instantly.
type Mock struct{}

func NewMock() *Mock { return &Mock{} }

func (m *Mock) ID() string      { return "mock" }
func (m *Mock) Name() string    { return "Sandbox" }
func (m *Mock) EnvKey() string  { return "" } // never locked
func (m *Mock) DocsURL() string { return "" }

func (m *Mock) Models() []ModelSpec {
	styleParam := ParamSpec{
		Key: "style", Label: "Style", Kind: ParamSelect, Default: "cinematic",
		Options: opts("cinematic", "Cinematic", "editorial", "Editorial", "anime", "Anime", "product", "Product"),
	}
	return []ModelSpec{
		{
			ID: "mock/still", Name: "Sandbox Still", Modality: ModalityImage, CreditCost: 1, Featured: true,
			Description: "Photoreal stills. Runs without any API key so you can try the studio immediately.",
			Tags:        []string{"image", "no key needed"},
			Params:      []ParamSpec{aspectParam("1:1"), styleParam, negativePromptParam, seedParam},
			RefImages:   3,
		},
		{
			ID: "mock/still-xl", Name: "Sandbox Still XL", Modality: ModalityImage, CreditCost: 2,
			Description: "Higher resolution stills with more detail retention.",
			Tags:        []string{"image", "no key needed"},
			Params:      []ParamSpec{aspectParam("16:9"), resolutionParam("1080p", "1080p", "1080p", "2k", "2K"), styleParam, seedParam},
			RefImages:   3,
		},
		{
			ID: "mock/motion", Name: "Sandbox Motion", Modality: ModalityVideo, CreditCost: 5, Featured: true,
			Description: "Text-to-video with camera motion. Simulates a real queue so progress is visible.",
			Tags:        []string{"video", "no key needed"},
			Params: []ParamSpec{
				aspectParam("16:9"), durationParam(5), resolutionParam("720p"),
				{Key: "camera", Label: "Camera move", Kind: ParamSelect, Default: "static",
					Options: opts("static", "Static", "dolly_in", "Dolly in", "orbit", "Orbit", "crane_up", "Crane up", "fpv", "FPV drone")},
				negativePromptParam,
			},
			RefImages: 1,
		},
		{
			ID: "mock/motion-pro", Name: "Sandbox Motion Pro", Modality: ModalityVideo, CreditCost: 12,
			Description: "Longer clips with first and last frame control.",
			Tags:        []string{"video", "no key needed"},
			Params:      []ParamSpec{aspectParam("21:9"), durationParam(10, "5", "5 seconds", "10", "10 seconds", "15", "15 seconds"), resolutionParam("1080p")},
			RefImages:   2,
		},
		{
			ID: "mock/voice", Name: "Sandbox Voice", Modality: ModalityAudio, CreditCost: 1,
			Description: "Text to speech for voiceover drafts.",
			Tags:        []string{"audio", "no key needed"},
			Params: []ParamSpec{
				{Key: "voice", Label: "Voice", Kind: ParamSelect, Default: "narrator",
					Options: opts("narrator", "Narrator", "warm", "Warm", "bright", "Bright")},
			},
		},
	}
}

// Submit encodes the finish time into the external id, so Poll needs no shared
// state and the worker can be restarted without losing in-flight jobs.
func (m *Mock) Submit(_ context.Context, req SubmitRequest) (SubmitResult, error) {
	work := 6 * time.Second
	if req.Model.Modality == ModalityVideo {
		work = 14 * time.Second
	}
	if req.Model.Modality == ModalityAudio {
		work = 4 * time.Second
	}
	seed := hash(req.Prompt + req.Model.ID)
	id := fmt.Sprintf("mock_%d_%s", time.Now().Add(work).UnixMilli(), seed[:12])
	return SubmitResult{ExternalID: id}, nil
}

func (m *Mock) Poll(_ context.Context, req PollRequest) (PollResult, error) {
	parts := strings.Split(req.ExternalID, "_")
	if len(parts) != 3 {
		return PollResult{Failed: true, Error: "malformed sandbox job id"}, nil
	}
	finishMS, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return PollResult{Failed: true, Error: "malformed sandbox job id"}, nil
	}
	seed := parts[2]

	remaining := time.Until(time.UnixMilli(finishMS))
	if remaining > 0 {
		return PollResult{Progress: progressFor(remaining, req.Model.Modality)}, nil
	}
	return PollResult{Done: true, Progress: 100, Assets: m.assets(req.Model, seed)}, nil
}

func (m *Mock) assets(model ModelSpec, seed string) []ResultAsset {
	w, h := 1024, 1024
	switch model.Modality {
	case ModalityVideo:
		w, h = 1280, 720
	case ModalityImage:
		w, h = 1024, 768
	}

	switch model.Modality {
	case ModalityVideo:
		clip := sampleClips[int(hashInt(seed))%len(sampleClips)]
		return []ResultAsset{{
			Kind: "video", URL: clip, Width: w, Height: h, DurationMS: 10000,
			ThumbnailURL: placeholder(seed, w, h),
		}}
	case ModalityAudio:
		return []ResultAsset{{Kind: "audio", URL: sampleAudio, DurationMS: 6000}}
	default:
		out := make([]ResultAsset, 0, 2)
		for i := 0; i < 2; i++ {
			s := fmt.Sprintf("%s-%d", seed, i)
			out = append(out, ResultAsset{Kind: "image", URL: placeholder(s, w, h), ThumbnailURL: placeholder(s, 512, 512), Width: w, Height: h})
		}
		return out
	}
}

// Public placeholder media. Swapped out entirely once a real provider key is
// present -- nothing else in the app knows these exist.
var sampleClips = []string{
	"https://storage.googleapis.com/gtv-videos-bucket/sample/ForBiggerJoyrides.mp4",
	"https://storage.googleapis.com/gtv-videos-bucket/sample/ForBiggerBlazes.mp4",
	"https://storage.googleapis.com/gtv-videos-bucket/sample/ForBiggerEscapes.mp4",
	"https://storage.googleapis.com/gtv-videos-bucket/sample/ForBiggerFun.mp4",
}

const sampleAudio = "https://storage.googleapis.com/gtv-videos-bucket/sample/ElephantsDream.mp4"

func placeholder(seed string, w, h int) string {
	return fmt.Sprintf("https://picsum.photos/seed/%s/%d/%d", seed, w, h)
}

func progressFor(remaining time.Duration, mod Modality) int {
	total := 6 * time.Second
	if mod == ModalityVideo {
		total = 14 * time.Second
	}
	if mod == ModalityAudio {
		total = 4 * time.Second
	}
	done := float64(total-remaining) / float64(total) * 100
	if done < 3 {
		done = 3
	}
	if done > 97 {
		done = 97
	}
	return int(done)
}

func hash(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])
}

func hashInt(s string) uint32 {
	sum := sha256.Sum256([]byte(s))
	return binary.BigEndian.Uint32(sum[:4])
}
