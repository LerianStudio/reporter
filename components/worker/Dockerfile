# syntax=docker/dockerfile:1.4

# --- Build stage ---
  FROM --platform=$BUILDPLATFORM golang:1.23-alpine AS builder

  # Install git and garble
  RUN apk add --no-cache git && \
      go install mvdan.cc/garble@latest
  
  # Set workdir to root of monorepo
  WORKDIR /plugin-smart-templates
  
  # Copy Go module files
  COPY go.mod go.sum ./
  
  # Use BuildKit secret to access private modules
  RUN --mount=type=secret,id=github_token \
      GITHUB_TOKEN=$(cat /run/secrets/github_token) && \
      printf "machine github.com\nlogin x-oauth-basic\npassword %s\n" "$GITHUB_TOKEN" > ~/.netrc && \
      go mod download && \
      rm ~/.netrc
  
  # Copy full project
  COPY . .
  
  # Add garble to PATH and build binary
  ENV PATH="/root/go/bin:${PATH}"
  RUN CGO_ENABLED=0 GOOS=linux garble -tiny -literals build -a -tags netgo -ldflags='-w -extldflags "-static"' \
      -o /worker components/worker/cmd/app/main.go
  
  # --- Final minimal image ---
  FROM gcr.io/distroless/static-debian12
  
  # Copy obfuscated binary
  COPY --from=builder /worker /worker
  
  # Optional: expose the port your app listens on
  EXPOSE 4003
  
  # Define the default entrypoint
  ENTRYPOINT ["/worker"]
  