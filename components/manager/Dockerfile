# syntax=docker/dockerfile:1.4

FROM golang:1.23-alpine AS builder

# Instala git e garble (para módulos privados e ofuscação)
RUN apk add --no-cache git && \
    go install mvdan.cc/garble@latest

# Caminho raiz do projeto
WORKDIR /plugin-smart-templates

# Copia tudo do projeto
COPY . .

#List file and show go.mod
RUN ls -la /plugin-smart-templates && cat /plugin-smart-templates/go.mod

# Usa BuildKit secret para acessar módulos privados
RUN --mount=type=secret,id=github_token \
  GITHUB_TOKEN=$(cat /run/secrets/github_token) && \
  printf "machine github.com\nlogin x-oauth-basic\npassword %s\n" "$GITHUB_TOKEN" > ~/.netrc && \
  cd /plugin-smart-templates && go mod download && rm ~/.netrc

# Compila o binário com garble
RUN CGO_ENABLED=0 GOOS=linux garble -tiny -literals build -a -tags netgo \
  -ldflags='-w -extldflags "-static"' \
  -o /manager ./components/manager/cmd/app/main.go

# Imagem final leve
FROM alpine:latest
COPY --from=builder /manager /manager

EXPOSE 4004
ENTRYPOINT ["/manager"]