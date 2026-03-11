.PHONY: build clean deploy local test lint

BINARY_DIR=bin
FUNCTIONS=auth listings ai

build:
	@mkdir -p $(BINARY_DIR)
	@for func in $(FUNCTIONS); do \
		echo "Building $$func..."; \
		mkdir -p $(BINARY_DIR)/$$func; \
		GOOS=linux GOARCH=arm64 go build -tags lambda.norpc -o $(BINARY_DIR)/$$func/bootstrap ./cmd/$$func/; \
	done
	@echo "Build complete."

clean:
	rm -rf $(BINARY_DIR)

deploy: build
	sam deploy --guided

local:
	sam local start-api

test:
	go test ./...

lint:
	golangci-lint run

tidy:
	go mod tidy

vet:
	go vet ./...
