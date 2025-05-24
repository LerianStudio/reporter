FROM golang:1.23-alpine AS builder

WORKDIR /plugin-smart-templates

# Copy go mod and sum files to download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -tags netgo -ldflags '-w -extldflags "-static"' -o /worker components/worker/cmd/app/main.go

FROM alpine:latest

COPY --from=builder /worker /worker

# Command to run the executable
CMD ["/worker"]