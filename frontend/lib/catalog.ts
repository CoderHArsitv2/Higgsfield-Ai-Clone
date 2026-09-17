import type { PublicModel } from "./types";

const API = process.env.BACKEND_URL ?? "http://localhost:8080";

/**
 * The landing page is marketing: it must render even when the API is asleep
 * (Render free tier cold starts) or not deployed yet. So a failed fetch falls
 * back to a static snapshot rather than showing an empty or broken model wall.
 */
export async function getPublicCatalog(): Promise<{
  models: PublicModel[];
  live: boolean;
}> {
  try {
    const res = await fetch(`${API}/api/v1/catalog`, {
      next: { revalidate: 300 },
      signal: AbortSignal.timeout(4000),
    });
    if (!res.ok) throw new Error(String(res.status));
    const data = (await res.json()) as { models: PublicModel[] };
    if (!data.models?.length) throw new Error("empty catalogue");
    return { models: data.models, live: true };
  } catch {
    return { models: FALLBACK, live: false };
  }
}

const m = (
  name: string,
  provider_name: string,
  modality: PublicModel["modality"],
  description: string,
  enabled = false,
  featured = false,
): PublicModel => ({
  id: `${provider_name}/${name}`.toLowerCase().replace(/\s+/g, "-"),
  name,
  provider_name,
  modality,
  description,
  enabled,
  featured,
});

/** Mirrors the server catalogue. Kept in sync by hand; only used when offline. */
const FALLBACK: PublicModel[] = [
  m("Sandbox Still", "Sandbox", "image", "Photoreal stills. Runs without any API key.", true, true),
  m("Sandbox Motion", "Sandbox", "video", "Text-to-video with camera motion. No key needed.", true, true),
  m("Sandbox Voice", "Sandbox", "audio", "Text to speech for voiceover drafts.", true),
  m("Kling 3.0", "fal.ai", "video", "Photorealism with complex motion.", false, true),
  m("Seedance 2.0", "fal.ai", "video", "Native audio-video with synced lip-sync in one pass.", false, true),
  m("Veo 3.1", "Google AI", "video", "4K cinematic generation with native synced audio.", false, true),
  m("Sora 2", "OpenAI", "video", "World simulation with accurate physics.", false),
  m("Wan 2.7", "fal.ai", "video", "The speed and richness balance point.", false),
  m("Kling 2.6", "fal.ai", "video", "Fast, stable character animation.", false),
  m("Hailuo 02", "fal.ai", "video", "Strong prompt adherence on stylised motion.", false),
  m("Ray 3", "fal.ai", "video", "Fluid natural motion with scene coherence.", false),
  m("Hunyuan Video", "Replicate", "video", "Open-weights video model with strong motion.", false),
  m("LTX Video", "Replicate", "video", "Real-time class video generation.", false),
  m("Nano Banana Pro", "Google AI", "image", "Exceptional targeted edits that leave the frame alone.", false, true),
  m("GPT Image 1", "OpenAI", "image", "Best-in-class instruction following for edits.", false, true),
  m("FLUX 1.1 Pro Ultra", "fal.ai", "image", "High fidelity stills up to 4MP.", false),
  m("FLUX schnell", "Replicate", "image", "Four-step model for near-instant drafts.", false),
  m("Recraft V3", "fal.ai", "image", "Vector-aware generation with reliable in-image text.", false),
  m("Ideogram V3", "fal.ai", "image", "Typography and poster layouts.", false),
  m("Qwen Image", "fal.ai", "image", "Strong bilingual text rendering.", false),
  m("SDXL", "Replicate", "image", "Stable Diffusion XL with a large LoRA ecosystem.", false),
  m("Eleven Multilingual v2", "ElevenLabs", "audio", "Highest quality speech across 29 languages.", false, true),
  m("Eleven Turbo v2.5", "ElevenLabs", "audio", "Low latency speech for long scripts.", false),
  m("MusicGen", "Replicate", "audio", "Text to music for background beds.", false),
];
