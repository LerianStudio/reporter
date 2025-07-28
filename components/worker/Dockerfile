# syntax=docker/dockerfile:1.4

FROM golang:1.23-alpine AS builder

# Instala git e garble (para módulos privados e ofuscação)
RUN apk add --no-cache git && \
    go install mvdan.cc/garble@latest

WORKDIR /plugin-smart-templates

COPY ../.. .

# Usa BuildKit secret para acessar módulos privados
RUN --mount=type=secret,id=github_token \
  GITHUB_TOKEN=$(cat /run/secrets/github_token) && \
  printf "machine github.com\nlogin x-oauth-basic\npassword %s\n" "$GITHUB_TOKEN" > ~/.netrc && \
  go mod download && rm ~/.netrc

# Compila o binário com garble
RUN CGO_ENABLED=0 GOOS=linux garble -tiny -literals build -a -tags netgo \
  -ldflags='-w -extldflags "-static"' \
  -o /worker ./components/worker/cmd/app/main.go

# Imagem final leve com Alpine
FROM alpine:latest
COPY --from=builder /worker /worker

EXPOSE 4005
ENTRYPOINT ["/worker"]