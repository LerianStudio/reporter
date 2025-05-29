FROM golang:1.23-alpine AS builder

WORKDIR /plugin-smart-templates

COPY ../.. .

RUN CGO_ENABLED=0 GOOS=linux go build -a -tags netgo -ldflags '-w -extldflags "-static"' -o /app components/manager/cmd/app/main.go

FROM alpine:latest

COPY --from=builder /app /app

EXPOSE 4004

ENTRYPOINT ["/app"]

#FROM golang:1.23-alpine AS builder
#
#WORKDIR /plugin-smart-templates
#
## Copy go mod and sum files to download dependencies
#COPY go.mod go.sum ./
#
#RUN go mod download
#
## Copy the source code
#COPY . .
#
## Build the application
#RUN CGO_ENABLED=0 GOOS=linux go build -a -tags netgo -ldflags '-w -extldflags "-static"' -o /app components/manager/cmd/app/main.go
#
#FROM alpine:latest
#
#COPY --from=builder /app /app
#
#EXPOSE 4004
#
#ENTRYPOINT ["/app"]