# Render builds from this. Multi-stage so the shipped image has no toolchain.
FROM golang:1.25-alpine AS build

WORKDIR /src
RUN apk add --no-cache ca-certificates

# Copy manifests first so dependency layers cache across code-only changes.
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/api ./cmd/api

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata \
  && adduser -D -u 10001 app
WORKDIR /app

COPY --from=build /out/api /app/api

# Generated media is written here. On Render this is ephemeral unless a disk is
# attached; mount one at /app/.media if you need results to survive a restart.
RUN mkdir -p /app/.media && chown -R app:app /app
USER app

ENV PORT=8080 MEDIA_DIR=/app/.media
EXPOSE 8080
CMD ["/app/api"]
