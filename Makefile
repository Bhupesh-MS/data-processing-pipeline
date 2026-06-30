.PHONY: format lint setup-hooks

# Format Go code
format:
	gofmt -s -w .
	go run golang.org/x/tools/cmd/goimports@latest -w .

# Lint Go code using golangci-lint
lint:
	go run github.com/golangci/golangci-lint/cmd/golangci-lint@latest run --fix ./...

# Setup git hooks to use the local .githooks directory
setup-hooks:
	git config core.hooksPath .githooks
	chmod +x .githooks/*
	@echo "Git hooks configured successfully."
