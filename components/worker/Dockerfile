# syntax=docker/dockerfile:1.4

FROM golang:1.23-alpine AS builder

# Instala garble e dependências
RUN apk add --no-cache git && \
    go install mvdan.cc/garble@latest

# Define diretório de trabalho principal do módulo
WORKDIR /plugin-smart-templates

# Copia arquivos de dependência antes de tudo
COPY ../../go.mod ../../go.sum ./

# Baixa as dependências (com autenticação para módulos privados)
RUN --mount=type=secret,id=github_token \
  GITHUB_TOKEN=$(cat /run/secrets/github_token) && \
  printf "machine github.com\nlogin x-oauth-basic\npassword %s\n" "$GITHUB_TOKEN" > ~/.netrc && \
  go mod download && \
  rm ~/.netrc

# Copia o restante do código
COPY ../.. .

# Garante que garble esteja no PATH
ENV PATH="/root/go/bin:${PATH}"

# Compila o worker com garble
RUN CGO_ENABLED=0 GOOS=linux garble -tiny -literals build -a -tags netgo \
  -ldflags='-w -extldflags "-static"' \
  -o /worker components/worker/cmd/app/main.go

# Imagem final mínima
FROM alpine:latest
COPY --from=builder /worker /worker
EXPOSE 4005
ENTRYPOINT ["/worker"]