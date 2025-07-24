# syntax=docker/dockerfile:1.4

# --- Build stage ---
  FROM golang:1.23-alpine AS builder

  # Install git and garble
  RUN apk add --no-cache git && \
      go install mvdan.cc/garble@latest
  
  WORKDIR /plugin-smart-templates
  
  # Copy dependency files separately to leverage Docker layer caching
  COPY go.mod go.sum ./
  
  # Use BuildKit secret for GitHub authentication (for private modules)
  RUN --mount=type=secret,id=github_token \
    GITHUB_TOKEN=$(cat /run/secrets/github_token) && \
    printf "machine github.com\nlogin x-oauth-basic\npassword %s\n" "$GITHUB_TOKEN" > ~/.netrc && \
    go mod download && \
    rm ~/.netrc
  
  # Copy the full source code
  COPY ../.. .
  
  # Obfuscate and build the binary using Garble
  ENV PATH="/root/go/bin:${PATH}"
  RUN CGO_ENABLED=0 GOOS=linux garble -tiny -literals build -a -tags netgo -ldflags '-w -extldflags "-static"' -o /app components/manager/cmd/app/main.go
  
  # --- Final minimal image ---
  FROM alpine:latest
  
  COPY --from=builder /app /app
  
  EXPOSE 4004
  
  ENTRYPOINT ["/app"]
  