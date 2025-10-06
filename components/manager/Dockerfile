# syntax=docker/dockerfile:1.4

FROM golang:1.25-alpine AS builder

# Install required packages: git is needed for private modules
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

# Copy only necessary source files for manager component
COPY pkg/ pkg/
COPY components/manager/ components/manager/

# Build the binary with garble for code obfuscation
# Exclude problematic packages using GOGARBLE with ! prefix
RUN GOGARBLE="!github.com/go-json-experiment/json,!github.com/chromedp/cdproto,!github.com/chromedp/chromedp" \
  CGO_ENABLED=0 GOOS=linux garble -tiny build -tags netgo \
  -ldflags='-w -s -extldflags "-static"' \
  -o /manager ./components/manager/cmd/app/main.go

# Stage 2: Final image with Chrome/Chromium for PDF generation
FROM alpine:latest

WORKDIR /app

# Install Chromium and all required dependencies for chromedp
RUN apk add --no-cache \
    chromium \
    nss \
    freetype \
    harfbuzz \
    ttf-freefont \
    ttf-dejavu \
    ttf-liberation \
    font-noto \
    font-noto-cjk \
    font-noto-emoji \
    font-noto-extra \
    && rm -rf /var/cache/apk/*

# Create a non-root user for security
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# Set Chrome/Chromium environment variables for headless operation
ENV CHROME_BIN=/usr/bin/chromium-browser
ENV CHROME_PATH=/usr/bin/chromium-browser
ENV CHROMEDP_SKIP_CHROMEDP_DOWNLOAD=true
ENV CHROMEDP_USE_SYSTEM_CHROME=true

# Copy the binary from the builder stage
COPY --from=builder /manager /app/manager

# Change ownership to non-root user
RUN chown -R appuser:appgroup /app

# Switch to non-root user
USER appuser

# Expose the application port
EXPOSE 4004

# Set the entrypoint to the manager binary
ENTRYPOINT ["/app/manager"]


