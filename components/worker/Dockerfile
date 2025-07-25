# syntax=docker/dockerfile:1.4

# --- Build stage ---
  FROM golang:1.23-alpine AS builder

  # Instala dependências
  RUN apk add --no-cache git && \
      go install mvdan.cc/garble@latest
  
  WORKDIR /plugin-smart-templates
  
  # Copia go.mod e go.sum da raiz do repositório
  COPY ../../go.mod ../../go.sum ./
  
  # Baixa dependências privadas usando BuildKit secret
  RUN --mount=type=secret,id=github_token \
      GITHUB_TOKEN=$(cat /run/secrets/github_token) && \
      printf "machine github.com\nlogin x-oauth-basic\npassword %s\n" "$GITHUB_TOKEN" > ~/.netrc && \
      go mod download && \
      rm ~/.netrc
  
  # Copia o restante do repositório
  COPY ../../ .
  
  # Compila com garble
  ENV PATH="/root/go/bin:${PATH}"
  RUN CGO_ENABLED=0 GOOS=linux garble -tiny -literals build -a -tags netgo \
      -ldflags='-w -extldflags "-static"' -o /worker components/worker/cmd/app/main.go
  
  # --- Final image ---
  FROM alpine:latest
  COPY --from=builder /worker /worker
  ENTRYPOINT ["/worker"]  