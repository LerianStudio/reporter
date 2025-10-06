# syntax=docker/dockerfile:1.4

FROM golang:1.25-alpine AS builder

# Install required packages: git is needed for private modules (e.g., from GitHub)
RUN apk add --no-cache git

# Install garble separately to improve layer caching build times
RUN go install mvdan.cc/garble@latest

# Set the working directory for the build
WORKDIR /plugin-smart-templates

# Copy only go.mod and go.sum first to cache dependencies
COPY go.mod go.sum ./

# Download Go modules using GitHub token from BuildKit secret
RUN --mount=type=secret,id=github_token \
  GITHUB_TOKEN=$(cat /run/secrets/github_token) && \
  printf "machine github.com\nlogin x-oauth-basic\npassword %s\n" "$GITHUB_TOKEN" > ~/.netrc && \
  go mod download && rm ~/.netrc

# Copy only necessary source files for worker component
COPY pkg/ pkg/
COPY components/worker/ components/worker/

# Build the binary with garble for code obfuscation
RUN CGO_ENABLED=0 GOOS=linux garble -tiny build -tags netgo \
  -ldflags='-w -extldflags "-static"' \
  -o /worker ./components/worker/cmd/app/main.go

# Stage 2: Final minimal image with only the binary
FROM alpine:latest

# Copy the obfuscated binary from the builder stage
COPY --from=builder /worker /worker

# Expose the application port
EXPOSE 4005

# Set the entrypoint to the worker binary
ENTRYPOINT ["/worker"]
