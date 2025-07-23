FROM golang:1.23-alpine AS builder

RUN apk add --no-cache git

WORKDIR /plugin-smart-templates

COPY go.mod go.sum ./

RUN --mount=type=secret,id=github_token \
  GITHUB_TOKEN=$(cat /run/secrets/github_token) && \
  printf "machine github.com\nlogin x-oauth-basic\npassword %s\n" "$GITHUB_TOKEN" > ~/.netrc && \
  go mod download && \
  rm ~/.netrc

COPY ../.. .

RUN CGO_ENABLED=0 GOOS=linux go build -a -tags netgo -ldflags '-w -extldflags "-static"' -o /app components/manager/cmd/app/main.go

FROM alpine:latest

COPY --from=builder /app /app

EXPOSE 4004

ENTRYPOINT ["/app"]