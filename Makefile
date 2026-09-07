SHELL := /bin/sh

NAME := terraform-provider-purelymail
VERSION ?= 0.1.0
DIST_DIR ?= dist

.PHONY: deps fmt validate test lint build docs package acceptance

deps:
	go mod tidy

fmt:
	gofmt -w $$(find . -name '*.go' -not -path './vendor/*')

validate:
	go vet ./...

test:
	go test ./...

lint:
	@if command -v golangci-lint >/dev/null 2>&1; then golangci-lint run ./...; else echo "golangci-lint not installed; skipping"; fi

build:
	go build ./...

docs:
	@echo "Provider documentation is maintained under docs/."

package:
	mkdir -p $(DIST_DIR)
	CGO_ENABLED=0 go build -trimpath -ldflags "-X main.version=$(VERSION)" -o $(DIST_DIR)/$(NAME) .
	tar -C $(DIST_DIR) -czf $(DIST_DIR)/$(NAME)_$(VERSION)_$$(go env GOOS)_$$(go env GOARCH).tar.gz $(NAME)

acceptance:
	TF_ACC=1 go test ./... -count=1 -run 'TestAcc'
