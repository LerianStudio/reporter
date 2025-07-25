# syntax=docker/dockerfile:1.4
FROM golang:1.23-alpine AS builder

RUN apk add --no-cache git && \
    go install mvdan.cc/garble@latest

# Usa diretório temporário para resolver as dependências
WORKDIR /tmp

COPY ../.. .

RUN --mount=type=secret,id=github_token \
  GITHUB_TOKEN=$(cat /run/secrets/github_token) && \
  printf "machine github.com\nlogin x-oauth-basic\npassword %s\n" "$GITHUB_TOKEN" > ~/.netrc && \
  cd /tmp && go mod download && rm ~/.netrc

# Agora sim, entra na pasta do projeto
WORKDIR /plugin-smart-templates

ENV PATH="/root/go/bin:${PATH}"

RUN CGO_ENABLED=0 GOOS=linux garble -tiny -literals build -a -tags netgo \
  -ldflags='-w -extldflags "-static"' \
  -o /manager components/manager/cmd/app/main.go

FROM alpine:latest
COPY --from=builder /manager /manager
EXPOSE 4004
ENTRYPOINT ["/manager"]