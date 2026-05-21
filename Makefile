# guitars-webapp Makefile
#
# Required environment variables for deploy/invalidate:
#   BUCKET  - S3 bucket name (e.g. the BucketName output from template.yaml)
#   DIST    - CloudFront distribution ID (DistributionId output from template.yaml)
#
# Required environment variables for build (passed through to Vite):
#   VITE_GUITARS_API_BASE_URL
#   VITE_GUITARS_BEARER_TOKEN   (optional; see README security section)

BUCKET ?=
DIST   ?=

.PHONY: install dev test lint build deploy invalidate clean

install:
	npm install

dev:
	npm run dev

test:
	npx vitest run

lint:
	npx tsc --noEmit

build:
	npm run build

deploy:
	@if [ -z "$(BUCKET)" ]; then echo "BUCKET is required"; exit 1; fi
	aws s3 sync dist/ s3://$(BUCKET) --delete

invalidate:
	@if [ -z "$(DIST)" ]; then echo "DIST is required"; exit 1; fi
	aws cloudfront create-invalidation --distribution-id $(DIST) --paths "/*"

clean:
	rm -rf dist node_modules
