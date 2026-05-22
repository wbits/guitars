#!/usr/bin/env bash
# LocalStack init hook: provisions the resources used by the GuitarCollection
# API when the LocalStack container reaches "ready" state.

set -euo pipefail

ENDPOINT="${AWS_ENDPOINT_URL:-http://localhost:4566}"
REGION="${AWS_DEFAULT_REGION:-us-east-1}"
TABLE_NAME="${GUITARS_TABLE:-Guitars}"
SECRET_NAME="${BEARER_SECRET_ID:-guitars/bearer-token}"
BEARER_TOKEN="${BEARER_TOKEN:-local-dev-token}"

export AWS_ACCESS_KEY_ID=test
export AWS_SECRET_ACCESS_KEY=test
export AWS_DEFAULT_REGION="${REGION}"

echo "Creating DynamoDB table ${TABLE_NAME} ..."
aws --endpoint-url="${ENDPOINT}" dynamodb create-table \
  --table-name "${TABLE_NAME}" \
  --attribute-definitions AttributeName=id,AttributeType=S \
  --key-schema AttributeName=id,KeyType=HASH \
  --billing-mode PAY_PER_REQUEST \
  >/dev/null 2>&1 || echo "  (table already exists)"

echo "Creating Secrets Manager secret ${SECRET_NAME} ..."
if ! aws --endpoint-url="${ENDPOINT}" secretsmanager describe-secret \
        --secret-id "${SECRET_NAME}" >/dev/null 2>&1; then
  aws --endpoint-url="${ENDPOINT}" secretsmanager create-secret \
    --name "${SECRET_NAME}" \
    --secret-string "${BEARER_TOKEN}" >/dev/null
else
  aws --endpoint-url="${ENDPOINT}" secretsmanager put-secret-value \
    --secret-id "${SECRET_NAME}" \
    --secret-string "${BEARER_TOKEN}" >/dev/null
fi

echo "LocalStack init complete."
echo "  table : ${TABLE_NAME}"
echo "  secret: ${SECRET_NAME}"
echo "  token : ${BEARER_TOKEN}"
