#!/usr/bin/env bash
# LocalStack init hook: provisions the resources used by the GuitarCollection
# API when the LocalStack container reaches "ready" state.

set -euo pipefail

ENDPOINT="${AWS_ENDPOINT_URL:-http://localhost:4566}"
REGION="${AWS_DEFAULT_REGION:-us-east-1}"
TABLE_NAME="${GUITARS_TABLE:-Guitars}"
MARKET_LOGS_TABLE="${MARKET_LOGS_TABLE:-MarketLogs}"
USER_PROFILES_TABLE="${USER_PROFILES_TABLE:-UserProfiles}"
GUITAR_ANALYSIS_TABLE="${GUITAR_ANALYSIS_TABLE:-GuitarAnalysis}"
ASSISTANT_USAGE_TABLE="${ASSISTANT_USAGE_TABLE:-AssistantUsage}"
SECRET_NAME="${BEARER_SECRET_ID:-guitars/bearer-token}"
BEARER_TOKEN="${BEARER_TOKEN:-local-dev-token}"
IMAGES_BUCKET="${IMAGES_BUCKET:-guitars-local}"

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

echo "Creating DynamoDB table ${MARKET_LOGS_TABLE} ..."
aws --endpoint-url="${ENDPOINT}" dynamodb create-table \
  --table-name "${MARKET_LOGS_TABLE}" \
  --attribute-definitions \
      AttributeName=id,AttributeType=S \
      AttributeName=guitarId,AttributeType=S \
      AttributeName=observedAt,AttributeType=S \
  --key-schema AttributeName=id,KeyType=HASH \
  --global-secondary-indexes '[
    {
      "IndexName": "guitarIdIndex",
      "KeySchema": [
        {"AttributeName": "guitarId", "KeyType": "HASH"},
        {"AttributeName": "observedAt", "KeyType": "RANGE"}
      ],
      "Projection": {"ProjectionType": "ALL"}
    }
  ]' \
  --billing-mode PAY_PER_REQUEST \
  >/dev/null 2>&1 || echo "  (table already exists)"

echo "Creating DynamoDB table ${USER_PROFILES_TABLE} ..."
aws --endpoint-url="${ENDPOINT}" dynamodb create-table \
  --table-name "${USER_PROFILES_TABLE}" \
  --attribute-definitions \
      AttributeName=userId,AttributeType=S \
      AttributeName=username,AttributeType=S \
  --key-schema AttributeName=userId,KeyType=HASH \
  --global-secondary-indexes '[
    {
      "IndexName": "usernameIndex",
      "KeySchema": [
        {"AttributeName": "username", "KeyType": "HASH"}
      ],
      "Projection": {"ProjectionType": "ALL"}
    }
  ]' \
  --billing-mode PAY_PER_REQUEST \
  >/dev/null 2>&1 || echo "  (table already exists)"

echo "Creating DynamoDB table ${ASSISTANT_USAGE_TABLE} ..."
aws --endpoint-url="${ENDPOINT}" dynamodb create-table \
  --table-name "${ASSISTANT_USAGE_TABLE}" \
  --attribute-definitions AttributeName=pk,AttributeType=S \
  --key-schema AttributeName=pk,KeyType=HASH \
  --billing-mode PAY_PER_REQUEST \
  >/dev/null 2>&1 || echo "  (table already exists)"
aws --endpoint-url="${ENDPOINT}" dynamodb update-time-to-live \
  --table-name "${ASSISTANT_USAGE_TABLE}" \
  --time-to-live-specification "Enabled=true, AttributeName=expiresAt" \
  >/dev/null 2>&1 || true

echo "Creating DynamoDB table ${GUITAR_ANALYSIS_TABLE} ..."
aws --endpoint-url="${ENDPOINT}" dynamodb create-table \
  --table-name "${GUITAR_ANALYSIS_TABLE}" \
  --attribute-definitions AttributeName=guitarId,AttributeType=S \
  --key-schema AttributeName=guitarId,KeyType=HASH \
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

echo "Creating S3 bucket ${IMAGES_BUCKET} ..."
aws --endpoint-url="${ENDPOINT}" s3 mb "s3://${IMAGES_BUCKET}" \
  >/dev/null 2>&1 || echo "  (bucket already exists)"

CORS_CONFIG=$(cat <<EOF
{
  "CORSRules": [
    {
      "AllowedHeaders": ["*"],
      "AllowedMethods": ["GET", "PUT", "HEAD"],
      "AllowedOrigins": [
        "http://localhost:5173",
        "http://127.0.0.1:5173",
        "http://localhost:5174",
        "http://127.0.0.1:5174"
      ],
      "ExposeHeaders": ["ETag"],
      "MaxAgeSeconds": 3600
    }
  ]
}
EOF
)
aws --endpoint-url="${ENDPOINT}" s3api put-bucket-cors \
  --bucket "${IMAGES_BUCKET}" \
  --cors-configuration "${CORS_CONFIG}" >/dev/null

echo "LocalStack init complete."
echo "  table : ${TABLE_NAME}"
echo "  market: ${MARKET_LOGS_TABLE}"
echo "  profiles: ${USER_PROFILES_TABLE}"
echo "  analysis: ${GUITAR_ANALYSIS_TABLE}"
echo "  secret: ${SECRET_NAME}"
echo "  token : ${BEARER_TOKEN}"
echo "  bucket: ${IMAGES_BUCKET}"
