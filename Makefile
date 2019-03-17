default: lint test
.PHONY: default

ci: setup lint test
.PHONY: ci

setup:
	@echo "--- setup install deps"
	@GO111MODULE=off go get -v -u github.com/golangci/golangci-lint/cmd/golangci-lint
.PHONY: setup

lint:
	@echo "--- lint all the things"
	@golangci-lint run
.PHONY: lint

test:
	@echo "--- test all the things"
	@go test -cover ./...
.PHONY: test

mocks:
	@echo "--- mock all the things"
	mockery -all -dir ./pkg/launcher
	mockery -all -dir ./pkg/cwlogs
.PHONY: mocks