# syntax=docker/dockerfile:1.4

# --- Build stage ---
  FROM golang:1.23-alpine AS builder

  # Instala o garble e dependências
  RUN apk add --no-cache git && \
      go install mvdan.cc/garble@latest
  
  # Diretório de trabalho do builder
  WORKDIR /plugin-smart-templates
  
  # Copia go.mod e go.sum a partir do contexto do build (manager)
  COPY ../../go.mod ../../go.sum ./
  
  # Usa BuildKit secret para acessar repositórios privados
  RUN --mount=type=secret,id=github_token \
    GITHUB_TOKEN=$(cat /run/secrets/github_token) && \
    printf "machine github.com\nlogin x-oauth-basic\npassword %s\n" "$GITHUB_TOKEN" > ~/.netrc && \
    go mod download && \
    rm ~/.netrc
  
  # Copia o restante do projeto (a raiz do repo)
  COPY ../../ .
  
  # Builda com garble
  ENV PATH="/root/go/bin:${PATH}"
  RUN CGO_ENABLED=0 GOOS=linux garble -tiny -literals build -a -tags netgo \
      -ldflags='-w -extldflags "-static"' -o /manager components/manager/cmd/app/main.go
  
  # --- Final minimal image ---
  FROM alpine:latest
  COPY --from=builder /manager /manager
  EXPOSE 4004
  ENTRYPOINT ["/manager"]  