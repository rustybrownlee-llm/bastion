#!/bin/bash
set -e

BASE_URL="http://localhost:8080"

echo "=== Testing Bastion POC Authentication ==="
echo ""

echo "1. Creating user..."
CREATE_RESPONSE=$(curl -s -X POST "$BASE_URL/api/v1/users" \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"secret123"}')
echo "$CREATE_RESPONSE"
echo ""

echo "2. Logging in..."
LOGIN_RESPONSE=$(curl -s -X POST "$BASE_URL/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"secret123"}')
echo "$LOGIN_RESPONSE"
echo ""

ACCESS_TOKEN=$(echo "$LOGIN_RESPONSE" | grep -o '"access_token":"[^"]*"' | cut -d'"' -f4)
REFRESH_TOKEN=$(echo "$LOGIN_RESPONSE" | grep -o '"refresh_token":"[^"]*"' | cut -d'"' -f4)

echo "3. Accessing protected endpoint (GET /users/me)..."
ME_RESPONSE=$(curl -s "$BASE_URL/api/v1/users/me" \
  -H "Authorization: Bearer $ACCESS_TOKEN")
echo "$ME_RESPONSE"
echo ""

echo "4. Refreshing access token..."
REFRESH_RESPONSE=$(curl -s -X POST "$BASE_URL/api/v1/auth/refresh" \
  -H "Content-Type: application/json" \
  -d "{\"refresh_token\":\"$REFRESH_TOKEN\"}")
echo "$REFRESH_RESPONSE"
echo ""

NEW_ACCESS_TOKEN=$(echo "$REFRESH_RESPONSE" | grep -o '"access_token":"[^"]*"' | cut -d'"' -f4)

echo "5. Using new access token..."
ME_RESPONSE2=$(curl -s "$BASE_URL/api/v1/users/me" \
  -H "Authorization: Bearer $NEW_ACCESS_TOKEN")
echo "$ME_RESPONSE2"
echo ""

echo "6. Logging out..."
curl -s -X POST "$BASE_URL/api/v1/auth/logout" \
  -H "Authorization: Bearer $NEW_ACCESS_TOKEN"
echo "Logged out successfully"
echo ""

echo "7. Attempting to refresh with logged out session (should fail)..."
REFRESH_FAIL=$(curl -s -X POST "$BASE_URL/api/v1/auth/refresh" \
  -H "Content-Type: application/json" \
  -d "{\"refresh_token\":\"$REFRESH_TOKEN\"}")
echo "$REFRESH_FAIL"
echo ""

echo "=== All tests completed ==="
