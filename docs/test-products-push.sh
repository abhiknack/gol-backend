#!/bin/bash

# Test script for POST /api/v1/products/push endpoint
# Usage: ./test-products-push.sh [base_url]
# Example: ./test-products-push.sh http://localhost:8080

BASE_URL="${1:-http://localhost:8080}"
ENDPOINT="${BASE_URL}/api/v1/products/push"

echo "Testing POST ${ENDPOINT}"
echo "================================"

curl -X POST "${ENDPOINT}" \
  -H "Content-Type: application/json" \
  -d @docs/API-PRODUCTS-PUSH-EXAMPLE.json \
  | jq '.'

echo ""
echo "================================"
echo "Test completed"
