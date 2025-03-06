VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -X gatehill.io/imposter/internal/config.version=$(VERSION)

.PHONY: build
build:
	go build -tags lambda.norpc -ldflags "$(LDFLAGS)" -o imposter

.PHONY: fmt
fmt:
	go fmt ./... 

.PHONY: run
run:
	go run -tags lambda.norpc -ldflags "$(LDFLAGS)" ./main.go $(filter-out $@,$(MAKECMDGOALS))

.PHONY: test
test:
	go test ./... 

.PHONY: coverage
coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

.PHONY: coverage-html
coverage-html: coverage
	go tool cover -html=coverage.out -o coverage.html 
