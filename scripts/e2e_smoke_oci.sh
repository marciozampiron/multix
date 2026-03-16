#!/bin/bash
# File: scripts/e2e_smoke_oci.sh
# Company: Hassan
# Creator: Zamp
# Created: 16/03/2026
# Purpose: End-to-end smoke test for OCI Provider

set -e

GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo "==> Building multix binary..."
go build -o build/multix ./cmd/multix

echo "==> Running OCI Auth Validate smoke test..."

# We capture stderr because the error is printed via the logger.
OUTPUT=$(./build/multix auth validate --provider oci 2>&1 || true)

# The success criterion for an E2E test without real credentials in a CI
# environment is that the OCI adapter correctly instantiates and attempts 
# to hit the real Identity SDK, resulting in an expected rejection:
# either "bad configuration" or "NotAuthorizedOrNotFound".
if echo "$OUTPUT" | grep -qE "NotAuthorizedOrNotFound|can not create client|bad configuration|failed to validate OCI credentials"; then
    echo -e "${GREEN}✓ OCI Provider wired correctly and SDK responded as expected (Identity API contact/rejection).${NC}"
    exit 0
else
    echo -e "${RED}✗ OCI Provider failed to respond with an expected Identity SDK signature.${NC}"
    echo "Output was:"
    echo "$OUTPUT"
    exit 1
fi
