# syntax=docker/dockerfile:1.4
FROM golang:1.23-alpine AS builder

# Instala o garble para ofuscar o binário
RUN apk add --no-cache git && \
    go install mvdan.cc/garble@latest

# Define diretório de trabalho no container
WORKDIR /plugin-smart-templates

# Copia os arquivos de dependência (precisam estar em components/worker)
COPY go.mod .
COPY go.sum .

# Baixa os módulos Go, com autenticação via token
RUN --mount=type=secret,id=github_token \
  GITHUB_TOKEN=$(cat /run/secrets/github_token) && \
  printf "machine github.com\nlogin x-oauth-basic\npassword %s\n" "$GITHUB_TOKEN" > ~/.netrc && \
  go mod download && \
  rm ~/.netrc

# Copia o restante do código (context: components/worker)
COPY . .

# Adiciona o garble ao PATH
ENV PATH="/root/go/bin:${PATH}"

# Compila o binário ofuscado
RUN CGO_ENABLED=0 GOOS=linux garble -tiny -literals build -a -tags netgo \
  -ldflags='-w -extldflags "-static"' \
  -o /worker ./cmd/app/main.go

# Stage final com Alpine
FROM alpine:latest
COPY --from=builder /worker /worker
EXPOSE 4005
ENTRYPOINT ["/worker"]
