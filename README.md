# Aperture

A rebuild of [higgsfield.ai](https://higgsfield.ai) — a multi-model generative
media studio. Next.js frontend, Go backend, Auth0 for identity.

It runs end to end with **no API keys at all**: a sandbox provider simulates a
real queue so the whole loop — auth, job submission, polling, gallery, credits —
is exercisable before you add a single credential.

---

## What it does

Every model vendor sits behind one interface. The catalogue lists all 34 models
regardless of what you have keys for; a model unlocks when a credential exists,
either the platform's or the signed-in user's own.

| | |
| --- | --- |
| **Frontend** | Next.js 15 (App Router), TypeScript, Tailwind v4, GSAP |
| **Backend** | Go 1.25, gin, gorm, Postgres |
| **Auth** | Auth0 OAuth only — no passwords, no local accounts |
| **Deploy** | Vercel (frontend) + Render (backend), production only |

---

## Departures from the original

The brief was to fix what's wrong with higgsfield.ai, not to copy it.

**The landing page is six sections, not ten.** The original runs five separate
product-announcement blocks — API, Genjutsu, Motion Designer, Effects, GPT Image
— each with its own CTA, all competing for the same attention. Those collapse
into a single **model wall**: the breadth *is* the pitch, so showing 34 models at
once says more than five banners do.

**Locked models are shown, not hidden.** A lock is an invitation to add a key.
Hiding them would make the product look smaller than it is.

**Motion is used once per section, deliberately.** A masked character reveal on
the hero, a pinned scrub through the product loop, a horizontal showcase driven
by vertical scroll. Everything respects `prefers-reduced-motion`.

---

## Architecture

```
backend/
  cmd/api/              entrypoint, wiring, graceful shutdown
  internal/
    config/             env parsing, fails fast on bad config
    controllers/        HTTP layer only — no SQL, no provider calls
    middleware/         Auth0 token verification, CORS, logging
    models/             gorm models
    provider/           the vendor seam (see below)
    routes/             route table
    services/           generation lifecycle, BYOK keys, polling worker
    storage/            re-hosts provider output that arrives as raw bytes
  pkg/
    apierr/  cryptox/  httpx/  jwtx/

frontend/
  app/                  routes; /api/proxy attaches the access token server-side
  components/landing/   the six landing sections
  components/studio/    workspace, model picker, BYOK manager
  lib/                  auth0 client, typed API client, GSAP setup
```

### The provider seam

Adding a vendor means adding one file that implements `provider.Provider` and
registering it in `main.go`. Nothing else changes.

```go
type Provider interface {
    ID() string
    Name() string
    EnvKey() string   // "" means no credential needed
    DocsURL() string
    Models() []ModelSpec
    Submit(ctx, SubmitRequest) (SubmitResult, error)
    Poll(ctx, PollRequest) (PollResult, error)
}
```

The registry resolves, per user, whether each model is unlocked by a platform
key (`server`), the user's own key (`byok`), or not at all (`locked`).

Models declare their own parameters, and the frontend renders controls from that
spec — so a new model with new options needs no frontend change.

### Jobs

Generations are a **database-backed queue**, not in-memory. A deploy or crash
mid-generation doesn't lose work the user already paid credits for; the worker
picks the same rows back up on restart. Failures refund automatically.

Providers that finish synchronously (OpenAI images, ElevenLabs speech) complete
inside `Submit` and never enter the polling path.

---

## Running locally

**Prerequisites:** Go 1.25+, Node 20+, Docker (for Postgres), an Auth0 tenant.

### 1. Auth0

Create a **Regular Web Application** and an **API**:

| Setting | Value |
| --- | --- |
| Allowed Callback URLs | `http://localhost:3000/auth/callback` |
| Allowed Logout URLs | `http://localhost:3000` |
| API Identifier | e.g. `https://api.higgsfield-clone.local` — this is `AUTH0_AUDIENCE` |

The audience matters: without it Auth0 issues an opaque token the Go API cannot
verify. It must be a **custom API you create** (Auth0 → Applications → APIs →
Create API), not Auth0's own Management API (`https://<tenant>/api/v2/`) — that
one mints tokens for managing your Auth0 tenant, not for your service, and a
regular web application is not authorised to request it by default.

### Sign-in screen

Sign-in happens on our own `/login` screen, not Auth0's Universal Login. Each
button links to `/auth/login?connection=<id>`, which the SDK forwards to Auth0's
`/authorize`. Naming the connection makes Auth0 skip its chooser and send the
user straight to that provider — click Google and the next thing you see is
Google's account picker.

The whole screen is plain links, so it ships no client JavaScript.

Which buttons appear is env-driven, because a connection only works once it is
enabled in Auth0 (Authentication → Social) *and* switched on for the
application. A button for an unconfigured connection returns an Auth0 error, so
the default is the one connection a fresh tenant has on:

```bash
AUTH0_CONNECTIONS=google-oauth2                    # default
AUTH0_CONNECTIONS=google-oauth2,windowslive,apple  # once those are configured
```

| id | Button |
| --- | --- |
| `google-oauth2` | Google |
| `windowslive` | Microsoft |
| `apple` | Apple |
| `github` | GitHub |

`AUTH0_EMAIL_CONNECTION` controls the "Continue with Email" option; leave it
empty to hand off to Universal Login without naming a connection.

### Callback URL

The callback URL is resolved from environment, in this order:

| Variable | Effect |
| --- | --- |
| `AUTH0_REDIRECT_URI` | Full URL. Wins over everything. Use behind ngrok, a tunnel or a proxy, where the origin the browser sees differs from the one the server knows. |
| `APP_BASE_URL` | Origin(s) the app is served from; the callback path is appended. A comma-separated list is accepted. |
| `AUTH0_CALLBACK_PATH` | The path itself, if `/auth/callback` collides with one of your routes. |

Whatever it resolves to must be pasted **verbatim** into the Auth0 application's
Allowed Callback URLs. In development the resolved value is printed at startup:

```
[auth0] callback URL: https://....ngrok-free.app/auth/callback
```

**Auth0 only accepts `https` callback URLs** for non-localhost origins, so a
tunnel is the usual way to develop against a real tenant:

```bash
ngrok http 3000
# then set both, and add the same callback to Auth0:
#   APP_BASE_URL=https://<subdomain>.ngrok-free.app
#   AUTH0_REDIRECT_URI=https://<subdomain>.ngrok-free.app/auth/callback
```

A free ngrok subdomain changes on every restart, which is exactly why this
lives in environment rather than in code.

### 2. Database

```bash
docker compose up -d
```

### 3. Backend

```bash
cd backend
cp .env.example .env
# set AUTH0_DOMAIN, AUTH0_AUDIENCE, and:
openssl rand -hex 32   # -> ENCRYPTION_KEY
go run ./cmd/api
```

### 4. Frontend

```bash
cd frontend
cp .env.example .env.local
# set the Auth0 values, and:
openssl rand -hex 32   # -> AUTH0_SECRET
npm install && npm run dev
```

Open `http://localhost:3000`, sign in, and run a **Sandbox** model — those work
with no keys.

---

## Provider keys

Everything below is optional. The catalogue shows all models either way; keys
decide which ones run.

| Env var | Unlocks |
| --- | --- |
| `FAL_KEY` | Kling 3.0/2.6, Seedance, Veo, Sora, Wan, Hailuo, Ray 3, FLUX, Nano Banana Pro, Recraft, Ideogram, Qwen — **14 models. Best single key to add.** |
| `REPLICATE_API_TOKEN` | FLUX, SDXL, Kling, Hunyuan, LTX, MiniMax, MusicGen — 8 models |
| `OPENAI_API_KEY` | GPT Image 1, Sora 2 |
| `GEMINI_API_KEY` | Nano Banana Pro, Nano Banana, Veo 3.1 |
| `ELEVENLABS_API_KEY` | Eleven Multilingual v2, Turbo v2.5 |

**Bring your own key.** Any user can add their own key at `/settings/keys`. It
unlocks that provider for their account only, is AES-256-GCM encrypted before it
touches the database, is never returned to the browser, and jobs run on it spend
no platform credits.

---

## Tests

```bash
cd backend
go test ./...                                    # unit tests, no infra needed

docker compose up -d
TEST_DATABASE_URL="postgres://postgres:postgres@localhost:5433/higgsfield?sslmode=disable" \
  go test ./...                                  # adds the integration suite
```

The integration suite drives a real job through the full loop against Postgres:
create → queue → worker submit → poll → assets → credit accounting.

---

## Deployment

Production only — there is no staging environment, by design.

`.github/workflows/deploy-prod.yml` runs on push to `main`: a test gate, then
two parallel deploys. The Render step polls the deploy to a terminal state and
then health-checks the service, because a fire-and-forget deploy hook reports
green even when the deploy failed.

**Required secrets:**

| Secret | Used for |
| --- | --- |
| `RENDER_API_KEY`, `RENDER_SERVICE_ID`, `BACKEND_PUBLIC_URL` | Backend deploy + health check |
| `VERCEL_TOKEN`, `VERCEL_ORG_ID`, `VERCEL_PROJECT_ID` | Frontend deploy |

`render.yaml` is a blueprint for the service and database. Set `CORS_ORIGINS` to
the Vercel domain and `PUBLIC_BASE_URL` to the Render domain.

> **Note on media storage.** Providers that return raw bytes get re-hosted to
> local disk, which is ephemeral on Render unless you attach a disk at
> `/app/.media`. The `storage.Storage` interface is one method wide so swapping
> in S3/R2 is a constructor change.

---

## Brand assets

| File | Use |
| --- | --- |
| `frontend/app/icon.svg` | Favicon. Next serves it by convention. |
| `frontend/app/apple-icon.png` | Apple touch icon, 180×180. |
| `frontend/public/logo-mark.svg` | The mark on its own, used in both navs. |
| `frontend/public/logo.svg` | Mark plus wordmark, for README and social cards. |
| `frontend/public/logo-mark-mono.svg` | Transparent background, for light surfaces. |

The mark is an aperture iris, generated from its geometry rather than drawn by
hand, so blade count, twist and opening are parameters. Stroke weights are a
fraction of the viewBox, so it renders identically at 16px and 512px. A
filled-wedge construction was tried first and rejected: below about 32px the
blades merge and it reads as an asterisk.

## Showcase footage

The clips in `frontend/public/showcase/` are placeholder stock footage from
Pexels, transcoded to ~200-450KB silent loops (1.9MB total). Swap in your own by
keeping the same filenames — see
[`CREDITS.md`](frontend/public/showcase/CREDITS.md).

## Agent logs

`.agent-logs/` contains the full prompt-and-response record of this build,
captured automatically. See [CAPTURE-TEST.md](CAPTURE-TEST.md).
