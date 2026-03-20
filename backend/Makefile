.PHONY: build clean deploy deploy-prod deploy-production verify-prod verify-production local test lint tidy vet

BINARY_DIR=bin
FUNCTIONS=auth listings payment
STACK_NAME?=propti-backend
AWS_REGION?=ap-southeast-1
API_CUSTOM_DOMAIN_NAME?=api.propti.id
API_CUSTOM_DOMAIN_CERTIFICATE_ARN?=arn:aws:acm:ap-southeast-1:260317865867:certificate/a6625ab7-0527-4dbf-aa1e-f22a39a33e98

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

deploy-prod: test
	sam build
	sam deploy \
		--no-confirm-changeset \
		--no-fail-on-empty-changeset \
		--resolve-s3 \
		--stack-name $(STACK_NAME) \
		--region $(AWS_REGION) \
		--capabilities CAPABILITY_IAM CAPABILITY_NAMED_IAM \
		--parameter-overrides \
			Stage=production \
			ApiCustomDomainName=$(API_CUSTOM_DOMAIN_NAME) \
			ApiCustomDomainCertificateArn=$(API_CUSTOM_DOMAIN_CERTIFICATE_ARN)

deploy-production: deploy-prod

verify-prod:
	aws cloudformation describe-stacks \
		--stack-name $(STACK_NAME) \
		--region $(AWS_REGION) \
		--query 'Stacks[0].{Status:StackStatus,Updated:LastUpdatedTime,Outputs:Outputs}'

verify-production: verify-prod

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
