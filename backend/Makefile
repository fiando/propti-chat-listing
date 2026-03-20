.PHONY: build clean deploy deploy-prod deploy-production local test lint tidy vet

BINARY_DIR=bin
FUNCTIONS=auth listings payment

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

deploy-prod: build
	sam deploy \
		--no-confirm-changeset \
		--no-fail-on-empty-changeset \
		--resolve-s3 \
		--stack-name propti-backend \
		--region ap-southeast-1 \
		--capabilities CAPABILITY_IAM CAPABILITY_NAMED_IAM \
		--parameter-overrides \
			Stage=production \
			ApiCustomDomainName=api.propti.id \
			ApiCustomDomainCertificateArn=arn:aws:acm:ap-southeast-1:260317865867:certificate/a6625ab7-0527-4dbf-aa1e-f22a39a33e98

deploy-production: deploy-prod

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
