# Dependency tool versions
GOTESTSUM_VERSION := v1.13.0
GOLANGCLI_VERSION := v2.10.1

BUILD_DIR := build

.PHONY: check test lint clean tools

check: test lint

test:
	@mkdir -p $(BUILD_DIR)
	gotestsum --format=short-verbose -- -race -coverprofile=$(BUILD_DIR)/coverage.txt -covermode=atomic
	go tool cover -html=$(BUILD_DIR)/coverage.txt -o $(BUILD_DIR)/coverage.html

lint:
	golangci-lint run

clean:
	rm -rf $(BUILD_DIR)

tools:
	go install gotest.tools/gotestsum@$(GOTESTSUM_VERSION)
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCLI_VERSION)
