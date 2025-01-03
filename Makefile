# Project-specific variables
project_name = SOLSniffer
APP_VERSION = $(VERSION)

# Display help menu
.PHONY: help
help: ## This help dialog.
	@grep -F -h "##" $(MAKEFILE_LIST) | grep -F -v fgrep | sed -e 's/\\$$//' | sed -e 's/##//'

# Run the app locally
.PHONY: run-local
run-local: ## Run the app locally
	go run cmd/main.go

# Manage dependencies
.PHONY: requirements
requirements: ## Generate go.mod & go.sum files
	go mod tidy
	go mod vendor

# Clean up Go modules
.PHONY: clean-packages
clean-packages: ## Clean packages
	go clean -modcache

# Build the project binary
.PHONY: build
build: ## Build the project
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w -X main.version=$(APP_VERSION)" -a -installsuffix cgo -o bin/$(project_name) ./cmd/main.go

# Run tests
.PHONY: test
test: ## Run tests
	go clean -testcache
	go test -p 1 -v -race ./...

# Run security checks
.PHONY: security
security: ## Run security checks
	@echo "Running security checks..."
	go install github.com/securego/gosec/v2/cmd/gosec@latest
	gosec ./...

# Run both tests and security checks
.PHONY: check
check: ## Run tests and security checks
	make test
	make security

# Install dependencies
.PHONY: deps
deps: ## Install dependencies
	go mod tidy

# Clean build artifacts
.PHONY: clean
clean: ## Clean build artifacts
	@echo "Cleaning up..."
	rm -rf bin/

# Docker build example
.PHONY: docker-build
docker-build: ## Build Docker image
	docker build -t $(project_name):$(APP_VERSION) .

# Docker push example
.PHONY: docker-push
docker-push: ## Push Docker image
	docker push $(project_name):$(APP_VERSION)
