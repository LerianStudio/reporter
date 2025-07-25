# syntax=docker/dockerfile:1.4
FROM golang:1.23-alpine AS builder

RUN apk add --no-cache git && \
    go install mvdan.cc/garble@latest

WORKDIR /build

COPY ../../go.mod .
COPY ../../go.sum .

RUN --mount=type=secret,id=github_token \
  GITHUB_TOKEN=$(cat /run/secrets/github_token) && \
  printf "machine github.com\nlogin x-oauth-basic\npassword %s\n" "$GITHUB_TOKEN" > ~/.netrc && \
  go mod download && rm ~/.netrc

COPY ../.. .

ENV PATH="/root/go/bin:${PATH}"

RUN CGO_ENABLED=0 GOOS=linux garble -tiny -literals build -a -tags netgo \
  -ldflags='-w -extldflags "-static"' \
  -o /worker ./components/worker/cmd/app/main.go

FROM alpine:latest
COPY --from=builder /worker /worker
EXPOSE 4005
ENTRYPOINT ["/worker"]