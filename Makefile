BUILD_DIR        := build
BOOTSTRAP        := $(BUILD_DIR)/bootstrap
PACKAGED         := packaged.yaml
TEMPLATE         := template.yaml
S3_BUCKET        := $(S3_BUCKET)
STACK_NAME       := $(STACK_NAME)

# LocalStack defaults; override on the command line to point at a real AWS env.
LOCALSTACK_ENDPOINT ?= http://localhost:4566
AWS_REGION          ?= us-east-1
GUITARS_TABLE       ?= Guitars
BEARER_SECRET_ID    ?= guitars/bearer-token
BEARER_TOKEN        ?= local-dev-token
SAM                 ?= $(shell brew --prefix aws-sam-cli 2>/dev/null)/bin/sam

.DEFAULT_GOAL := help

## help: show available targets
.PHONY: help
help:
	@grep -E '^##' Makefile | sed -e 's/## //'

## test: run all unit tests
.PHONY: test
test:
	GOTOOLCHAIN=local go test ./...

## vet: static checks
.PHONY: vet
vet:
	GOTOOLCHAIN=local go vet ./...

## tidy: clean up go.mod/go.sum
.PHONY: tidy
tidy:
	GOTOOLCHAIN=local go mod tidy

## install: download go dependencies
.PHONY: install
install:
	GOTOOLCHAIN=local go mod download

## build: cross-compile the lambda binary as build/bootstrap (provided.al2)
.PHONY: build
build: clean
	mkdir -p $(BUILD_DIR)
	GOTOOLCHAIN=local CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
	    go build -tags lambda.norpc -o $(BOOTSTRAP) ./cmd/lambda

## clean: remove build artefacts
.PHONY: clean
clean:
	rm -rf $(BUILD_DIR) $(PACKAGED)

## localstack-up: start LocalStack (DynamoDB + Secrets Manager) in docker
.PHONY: localstack-up
localstack-up:
	docker-compose up -d localstack

## localstack-down: stop LocalStack
.PHONY: localstack-down
localstack-down:
	docker compose down

## localstack-init: re-run the LocalStack init script (table + secret)
.PHONY: localstack-init
localstack-init:
	GUITARS_TABLE=$(GUITARS_TABLE) \
	BEARER_SECRET_ID=$(BEARER_SECRET_ID) \
	BEARER_TOKEN=$(BEARER_TOKEN) \
	AWS_ENDPOINT_URL=$(LOCALSTACK_ENDPOINT) \
	AWS_DEFAULT_REGION=$(AWS_REGION) \
	./scripts/localstack-init.sh

## api: run the API locally via SAM CLI against LocalStack
.PHONY: api
api: build
	$(SAM) local start-api \
	    --docker-network guitars-net \
	    --parameter-overrides \
	        TableName=$(GUITARS_TABLE) \
	        BearerSecretName=$(BEARER_SECRET_ID) \
	    --env-vars env.local.json \
	    --container-env-vars container.local.json

## package: produce a CloudFormation package (requires S3_BUCKET)
.PHONY: package
package: test build
	$(SAM) package --template-file $(TEMPLATE) --s3-bucket $(S3_BUCKET) --output-template-file $(PACKAGED)

## deploy: deploy to AWS (requires S3_BUCKET and STACK_NAME)
.PHONY: deploy
deploy: package
	$(SAM) deploy --stack-name $(STACK_NAME) --template-file $(PACKAGED) --capabilities CAPABILITY_IAM --no-confirm-changeset --region $(AWS_REGION)
