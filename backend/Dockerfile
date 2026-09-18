# Render builds this. Multi-stage so the shipped image has no toolchain.
#
# The build context is the REPOSITORY ROOT, not backend/, because that is what
# Render uses by default and a context that has to be configured by hand is a
# context someone will forget -- the symptom is a cryptic
# `"/go.sum": not found` during COPY. Paths below are therefore repo-relative,
# and the root .dockerignore keeps the frontend out of the upload.
#
#   docker build -f backend/Dockerfile .
FROM golang:1.25-alpine AS build

WORKDIR /src
RUN apk add --no-cache ca-certificates

# Manifests first so the dependency layer caches across code-only changes.
COPY backend/go.mod backend/go.sum ./
RUN go mod download

COPY backend/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/api ./cmd/api

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata \
  && adduser -D -u 10001 app
WORKDIR /app

COPY --from=build /out/api /app/api

# Generated media is written here when no object store is configured. On Render
# that is ephemeral, so attach a disk at /app/.media or set STORAGE_BUCKET.
RUN mkdir -p /app/.media && chown -R app:app /app
USER app

ENV PORT=8080 MEDIA_DIR=/app/.media
EXPOSE 8080
CMD ["/app/api"]
