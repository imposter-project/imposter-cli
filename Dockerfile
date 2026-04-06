FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build a statically linked binary
RUN CGO_ENABLED=0 go build -tags lambda.norpc -ldflags="-s -w" -o imposter-cli

# Create an empty directory to use in the scratch stage
RUN mkdir /empty

# Final stage
FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app/imposter-cli /usr/local/bin/imposter
COPY --from=builder /empty /mocks
COPY --from=builder /empty /tmp

WORKDIR /mocks

ENV IMPOSTER_ENGINE=golang

EXPOSE 8080

ENTRYPOINT ["imposter"]
CMD ["--help"]
