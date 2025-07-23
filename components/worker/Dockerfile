FROM golang:1.23-alpine AS builder

RUN apk add --no-cache git

WORKDIR /plugin-smart-templates

# Copy go mod and sum files to download dependencies
COPY go.mod go.sum ./

RUN --mount=type=secret,id=github_token \
  GITHUB_TOKEN=$(cat /run/secrets/github_token) && \
  printf "machine github.com\nlogin x-oauth-basic\npassword %s\n" "$GITHUB_TOKEN" > ~/.netrc && \
  go mod download && \
  rm ~/.netrc

RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -tags netgo -ldflags '-w -extldflags "-static"' -o /worker components/worker/cmd/app/main.go

FROM alpine:latest

COPY --from=builder /worker /worker

# Command to run the executable
CMD ["/worker"]