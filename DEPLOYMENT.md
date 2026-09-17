# Deployment

Production is the only environment. Frontend on Vercel, backend and Postgres on
Render.

---

## How a deploy happens

| Trigger | What runs | Version recorded |
| --- | --- | --- |
| Push to `main` | test gate → Render + Vercel deploys | `main-<short-sha>` |
| Push a `v*` tag | test gate → deploys → GitHub Release | the tag, e.g. `v1.0.0` |
| Manual (`workflow_dispatch`) | same as a `main` push | `main-<short-sha>` |

A branch push does not match a tag filter and a tag push does not match a branch
filter, so tagging a commit already on `main` deploys it a second time,
deliberately, to stamp the released version. It never double-fires for one push.

The gate must pass before anything deploys: `gofmt`, `go vet`, `go test -race`
against a real Postgres, a Docker build of the image Render will build, and on
the frontend Prettier, `tsc`, ESLint and a full `next build`.

The release is published only after **both** deploys succeed, so a release tag
never points at something that failed to reach production.

### Verifying a deploy actually landed

The health check waits until `/health` reports the exact version just deployed:

```json
{"status":"ok","env":"production","version":"v1.0.0","commit":"9f3c1ab"}
```

Checking only for `status: ok` would pass against the instance that was already
running. `commit` comes from `RENDER_GIT_COMMIT`, which Render injects on every
build, so even an untagged deploy is traceable to a commit.

---

## One-time setup

### 1. Render — database

1. **New → Postgres**. Name `aperture-db`, pick a region, create.
2. Copy the **Internal Database URL** (not external — internal is faster and
   stays inside Render's network).

### 2. Render — web service

1. **New → Web Service**, connect the repo.
2. **Runtime: Docker**. **Dockerfile path** `backend/Dockerfile`,
   **Docker build context** `backend`.
3. **Health check path**: `/health`.
4. **Auto-Deploy: off.** GitHub Actions owns deploys so they are gated on tests;
   leaving it on means an untested push deploys itself.
5. Environment variables:

   | Key | Value |
   | --- | --- |
   | `APP_ENV` | `production` |
   | `DATABASE_URL` | the internal URL from step 1 |
   | `AUTH0_DOMAIN` | your tenant, no scheme, no trailing slash |
   | `AUTH0_AUDIENCE` | your API identifier — **identical** to the frontend's |
   | `ENCRYPTION_KEY` | `openssl rand -hex 32` — see the warning below |
   | `CORS_ORIGINS` | your Vercel URL |
   | `PUBLIC_BASE_URL` | your Render URL |
   | `FAL_KEY` etc. | any provider keys you want platform-wide |

   > **`ENCRYPTION_KEY` is not rotatable in place.** It encrypts users' stored
   > provider keys. Change it and every stored key becomes undecryptable and has
   > to be re-entered. Generate it once and keep it somewhere you will not lose.

6. Deploy once by hand so the service exists, then copy the **service id**
   (`srv-…`) from the URL.

### 3. Vercel

1. **Add New → Project**, import the repo.
2. **Root Directory: `frontend`.** Everything else is detected.
3. Environment variables, all for **Production**:

   | Key | Value |
   | --- | --- |
   | `AUTH0_DOMAIN` | your tenant |
   | `AUTH0_CLIENT_ID` / `AUTH0_CLIENT_SECRET` | from the Auth0 application |
   | `AUTH0_AUDIENCE` | identical to the backend's |
   | `AUTH0_SECRET` | `openssl rand -hex 32` |
   | `APP_BASE_URL` | your Vercel production URL |
   | `AUTH0_REDIRECT_URI` | `<vercel url>/auth/callback` |
   | `BACKEND_URL` | your Render URL |
   | `AUTH0_CONNECTIONS` | e.g. `google-oauth2` |

4. Deploy once so the project exists, then take the **org id** and **project
   id** from Project Settings → General.

### 4. Auth0

Add to the application's **Allowed Callback URLs**:
`https://<your-vercel-domain>/auth/callback`, and to **Allowed Logout URLs**:
`https://<your-vercel-domain>`. These are additive — keep the ngrok entry if you
still develop against it.

### 5. GitHub secrets

Settings → Secrets and variables → Actions:

| Secret | Where it comes from |
| --- | --- |
| `RENDER_API_KEY` | Render → Account Settings → API Keys |
| `RENDER_SERVICE_ID` | the `srv-…` id from step 2 |
| `BACKEND_PUBLIC_URL` | your Render URL, no trailing slash |
| `VERCEL_TOKEN` | Vercel → Account Settings → Tokens |
| `VERCEL_ORG_ID`, `VERCEL_PROJECT_ID` | from step 3 |

Optionally create a `production` environment (Settings → Environments) and add
required reviewers, which makes every production deploy wait for an approval.

---

## Cutting a release

```bash
git tag -a v1.0.0 -m "First release"
git push origin v1.0.0
```

That runs the gate, deploys both, stamps `v1.0.0` on the backend, and publishes
a GitHub Release with notes generated from the commits since the last tag.

Roll back by re-running the workflow from an earlier tag: Actions → Deploy
production → Run workflow → pick the tag.

---

## Known limitations

**Generated media on Render is ephemeral.** Providers that return raw bytes
(OpenAI images, Gemini, ElevenLabs) get re-hosted to local disk, which is wiped
on every deploy and restart. Attach a Render disk mounted at `/app/.media`, or
swap `storage.Storage` for S3/R2 — the interface is one method wide.

**The free Render tier sleeps.** The first request after idle takes ~30s, which
the landing page survives (it falls back to a static catalogue) but the studio
will feel broken. Use a paid instance for anything demoed live.

**AutoMigrate only ever adds.** It will not drop a column, narrow a type, or
remove an index you deleted. Renaming a struct field leaves the old column
behind holding data.
