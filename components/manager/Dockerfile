FROM golang:1.23-alpine AS builder

WORKDIR /plugin-template-engine

COPY ../.. .

RUN CGO_ENABLED=0 GOOS=linux go build -a -tags netgo -ldflags '-w -extldflags "-static"' -o /app components/manager/cmd/app/main.go

FROM alpine:latest

COPY --from=builder /app /app

EXPOSE 4004

ENTRYPOINT ["/app"]