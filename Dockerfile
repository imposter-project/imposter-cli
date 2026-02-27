FROM golang:1.23-alpine AS builder

WORKDIR /app

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN go build -tags lambda.norpc -o imposter-cli

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates curl openjdk17-jre docker-cli

WORKDIR /app

# Copy the binary from builder stage
COPY --from=builder /app/imposter-cli /usr/local/bin/imposter

# Create directory for output
RUN mkdir -p /output

EXPOSE 8080

# Default command shows help
CMD ["imposter", "--help"]
