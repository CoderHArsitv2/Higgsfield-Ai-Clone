package services

import (
	"testing"

	"github.com/coderHArsitv2/higgsfield-clone/backend/internal/provider"
)

func specWithParams() provider.ModelSpec {
	return provider.ModelSpec{
		ID: "test/model",
		Params: []provider.ParamSpec{
			{Key: "aspect_ratio", Default: "16:9"},
			{Key: "seed"},
		},
	}
}

// Anything the model did not declare must be dropped, so a crafted request
// cannot smuggle extra fields into an upstream provider call.
func TestNormaliseParamsDropsUndeclaredKeys(t *testing.T) {
	out := normaliseParams(specWithParams(), map[string]any{
		"aspect_ratio":     "1:1",
		"seed":             42,
		"webhook_url":      "https://attacker.example",
		"x-internal-admin": true,
	})
	if _, ok := out["webhook_url"]; ok {
		t.Fatal("undeclared key survived normalisation")
	}
	if _, ok := out["x-internal-admin"]; ok {
		t.Fatal("undeclared key survived normalisation")
	}
	if out["aspect_ratio"] != "1:1" {
		t.Fatalf("declared value lost: %v", out["aspect_ratio"])
	}
	if out["seed"] != 42 {
		t.Fatalf("declared value lost: %v", out["seed"])
	}
}

func TestNormaliseParamsAppliesDefaults(t *testing.T) {
	out := normaliseParams(specWithParams(), nil)
	if out["aspect_ratio"] != "16:9" {
		t.Fatalf("default not applied: %v", out["aspect_ratio"])
	}
	// seed has no default and was not supplied, so it should be absent rather
	// than sent upstream as a zero value.
	if _, ok := out["seed"]; ok {
		t.Fatal("param with no default and no value was included")
	}
}

func TestNormaliseParamsTreatsEmptyStringAsUnset(t *testing.T) {
	out := normaliseParams(specWithParams(), map[string]any{"aspect_ratio": ""})
	if out["aspect_ratio"] != "16:9" {
		t.Fatalf("empty string should fall back to the default, got %v", out["aspect_ratio"])
	}
}
