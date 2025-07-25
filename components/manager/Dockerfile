# syntax=docker/dockerfile:1.4

FROM golang:1.23-alpine AS builder

# Instala dependências e garble
RUN apk add --no-cache git && \
    go install mvdan.cc/garble@latest

# Copia apenas os arquivos necessários para baixar os módulos
WORKDIR /tmp
COPY go.mod go.sum ./

# Baixa os módulos privados usando token como segredo
RUN --mount=type=secret,id=github_token \
    GITHUB_TOKEN=$(cat /run/secrets/github_token) && \
    printf "machine github.com\nlogin x-oauth-basic\npassword %s\n" "$GITHUB_TOKEN" > ~/.netrc && \
    go mod download && \
    rm ~/.netrc

# Copia o código da aplicação e compila com garble
WORKDIR /plugin-smart-templates
COPY . .

ENV PATH="/root/go/bin:${PATH}"
RUN CGO_ENABLED=0 GOOS=linux garble -tiny -literals build -a -tags netgo \
    -ldflags='-w -extldflags "-static"' -o /manager components/manager/cmd/app/main.go

# Imagem final minimalista
FROM alpine:latest
COPY --from=builder /manager /manager
EXPOSE 4004
ENTRYPOINT ["/manager"]