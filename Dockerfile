FROM golang:1.23-alpine AS builder

WORKDIR /k8s-golang-addons-boilerplate

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -tags netgo -ldflags '-w -extldflags "-static"' -o /app cmd/app/main.go

FROM alpine:latest

COPY --from=builder /app /app

COPY --from=builder /k8s-golang-addons-boilerplate/migrations /migrations

EXPOSE 3005 7001

ENTRYPOINT ["/app"]