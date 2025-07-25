# syntax=docker/dockerfile:1.4

FROM golang:1.23-alpine AS builder

# Instala git e garble para obfuscação
RUN apk add --no-cache git && \
    go install mvdan.cc/garble@latest

# Etapa intermediária para garantir que go mod download funcione
WORKDIR /tmp

# Copia a raiz do repositório (ajustado para contexto externo ao Dockerfile)
COPY ../.. .

# Baixa as dependências com autenticação para módulos privados
RUN --mount=type=secret,id=github_token \
  GITHUB_TOKEN=$(cat /run/secrets/github_token) && \
  printf "machine github.com\nlogin x-oauth-basic\npassword %s\n" "$GITHUB_TOKEN" > ~/.netrc && \
  cd /tmp && go mod download && rm ~/.netrc

# Vai para o diretório correto do projeto
WORKDIR /tmp/plugin-smart-templates

# Garante que garble esteja no PATH
ENV PATH="/root/go/bin:${PATH}"

# Compila o binário ofuscado com garble
RUN CGO_ENABLED=0 GOOS=linux garble -tiny -literals build -a -tags netgo \
  -ldflags='-w -extldflags "-static"' \
  -o /manager components/manager/cmd/app/main.go

# Imagem final minimalista
FROM alpine:latest

COPY --from=builder /manager /manager

EXPOSE 4004
ENTRYPOINT ["/manager"]