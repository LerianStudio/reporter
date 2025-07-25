# syntax=docker/dockerfile:1.4

# --- Build stage ---
  FROM golang:1.23-alpine AS builder

  # Instala git e garble
  RUN apk add --no-cache git && \
      go install mvdan.cc/garble@latest
  
  # Define diretório de trabalho
  WORKDIR /plugin-smart-templates
  
  # Copia arquivos de dependência
  COPY go.mod go.sum ./
  
  # Usa o token do GitHub para baixar dependências privadas
  RUN --mount=type=secret,id=github_token \
    GITHUB_TOKEN=$(cat /run/secrets/github_token) && \
    printf "machine github.com\nlogin x-oauth-basic\npassword %s\n" "$GITHUB_TOKEN" > ~/.netrc && \
    go mod download && \
    rm ~/.netrc
  
  # Copia todo o código-fonte do projeto
  COPY . .
  
  # Adiciona o garble no PATH e faz o build com ofuscação
  ENV PATH="/root/go/bin:${PATH}"
  RUN CGO_ENABLED=0 GOOS=linux garble -tiny -literals build -a -tags netgo \
      -ldflags='-w -extldflags "-static"' \
      -o /manager components/manager/cmd/app/main.go
  
  # --- Final minimal image ---
  FROM alpine:latest
  
  # Copia o binário compilado
  COPY --from=builder /manager /manager
  
  # Expondo a porta padrão
  EXPOSE 4004
  
  # Executa o binário
  ENTRYPOINT ["/manager"]  