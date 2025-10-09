#!/bin/bash

# Test script for the sync command
# This script demonstrates how to use the sync command to pull data from an external hub3 v2 API

set -e

echo "=== Testing hub3-cli sync command ==="
echo ""

# Build the CLI tool
echo "Building hub3-cli..."
cd tools/cmd/index_narthex
go build -o hub3-cli .
echo "✓ Build successful"
echo ""

# Test 1: Show help
echo "=== Test 1: Show help ==="
./hub3-cli sync --help
echo ""

# Test 2: Dry-run with public endpoint (if available)
# Replace with your actual public endpoint
echo "=== Test 2: Dry-run test ==="
echo "To test with a real endpoint, run:"
echo ""
echo "  ./hub3-cli sync \\"
echo "    --url https://your-hub3-instance.org/api/search/v2 \\"
echo "    --source-org your-org-id \\"
echo "    --index test-index \\"
echo "    --query 'meta.spec:dataset-name' \\"
echo "    --batch-size 10 \\"
echo "    --dry-run"
echo ""

# Test 3: Full sync example (commented out - requires real endpoint and ES)
echo "=== Test 3: Full sync example (requires ES) ==="
echo "To run a full sync to local Elasticsearch, run:"
echo ""
echo "  ./hub3-cli sync \\"
echo "    --url https://your-hub3-instance.org/api/search/v2 \\"
echo "    --source-org your-org-id \\"
echo "    --index hub3-test \\"
echo "    --query 'meta.spec:dataset-name' \\"
echo "    --batch-size 100 \\"
echo "    --workers 4 \\"
echo "    --es-url http://localhost:9200"
echo ""

echo "=== Examples ==="
echo ""
echo "1. Sync specific dataset:"
echo "   hub3-cli sync --url https://data.hub3.org/api/search/v2 \\"
echo "     --source-org brabantcloud --index hub3-dev \\"
echo "     --query 'meta.spec:my-dataset'"
echo ""
echo "2. Sync with different target org:"
echo "   hub3-cli sync --url https://prod.hub3.org/api/search/v2 \\"
echo "     --source-org prod-org --target-org dev-org \\"
echo "     --index hub3-dev"
echo ""
echo "3. Dry run to test query:"
echo "   hub3-cli sync --url https://data.hub3.org/api/search/v2 \\"
echo "     --source-org brabantcloud --index test \\"
echo "     --query '*' --dry-run"
echo ""

echo "✓ All tests completed"
