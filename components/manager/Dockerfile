# syntax=docker/dockerfile:1.4

# --- Build stage ---
  FROM golang:1.23-alpine AS builder

  # Install git and Garble (for module fetching and binary obfuscation)
  RUN apk add --no-cache git && \
      go install mvdan.cc/garble@latest
  
  # Set working directory
  WORKDIR /plugin-smart-templates
  
  # Copy dependency files first to optimize layer caching
  COPY go.mod go.sum ./
  
  # Use BuildKit secret to authenticate with GitHub and download private modules
  RUN --mount=type=secret,id=github_token \
    GITHUB_TOKEN=$(cat /run/secrets/github_token) && \
    printf "machine github.com\nlogin x-oauth-basic\npassword %s\n" "$GITHUB_TOKEN" > ~/.netrc && \
    go mod download && \
    rm ~/.netrc
  
  # Copy the entire project source code
  COPY ../.. .
  
  # Ensure Garble is in PATH and build the binary with obfuscation
  ENV PATH="/root/go/bin:${PATH}"
  RUN CGO_ENABLED=0 GOOS=linux garble -tiny -literals build -a -tags netgo -ldflags='-w -extldflags "-static"' \
    -o /manager components/manager/cmd/app/main.go
  
  # --- Final minimal image ---
  FROM alpine:latest
  
  # Copy the compiled and obfuscated binary
  COPY --from=builder /manager /manager
  
  # Expose the default application port
  EXPOSE 4004
  
  # Set the container entrypoint
  ENTRYPOINT ["/manager"]
  