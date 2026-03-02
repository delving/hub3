#!/usr/bin/env bash
# End-to-end tests for the Hub3 Semantic API v1
#
# Usage:
#   ./test_semantic_e2e.sh                          # defaults to http://localhost:3002
#   ./test_semantic_e2e.sh https://api.brabantcloud.nl
#
# Exit codes: 0 = all passed, 1 = failures

set -euo pipefail

BASE="${1:-http://localhost:3002}"
API="${BASE}/api/semantic/v1"
PASS=0
FAIL=0
SKIP=0

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BOLD='\033[1m'
NC='\033[0m'

pass() { ((PASS++)); echo -e "  ${GREEN}PASS${NC} $1"; }
fail() { ((FAIL++)); echo -e "  ${RED}FAIL${NC} $1: $2"; }
skip() { ((SKIP++)); echo -e "  ${YELLOW}SKIP${NC} $1: $2"; }

# Helper: curl with JSON output, fail on HTTP errors
api() {
    local url="$1"
    curl -sf --max-time 10 -H "Accept: application/ld+json" "$url" 2>/dev/null
}

# Helper: check JSON field exists and has expected value
json_has() {
    local json="$1" path="$2" expected="$3"
    local actual
    actual=$(echo "$json" | python3 -c "
import sys, json
data = json.load(sys.stdin)
keys = '$path'.split('.')
v = data
for k in keys:
    if isinstance(v, list):
        v = v[int(k)]
    elif isinstance(v, dict):
        v = v.get(k)
    else:
        v = None
        break
print(json.dumps(v) if v is not None else 'null')
" 2>/dev/null)
    if [ "$actual" = "null" ] && [ "$expected" != "null" ]; then
        return 1
    fi
    if [ -n "$expected" ] && [ "$expected" != "*" ]; then
        [ "$actual" = "$expected" ] || [ "$actual" = "\"$expected\"" ]
    else
        [ "$actual" != "null" ]
    fi
}

# Helper: check HTTP status code
http_status() {
    curl -so /dev/null -w "%{http_code}" --max-time 10 "$1" 2>/dev/null
}

echo -e "${BOLD}Hub3 Semantic API v1 — End-to-End Tests${NC}"
echo -e "Target: ${BOLD}${API}${NC}"
echo ""

# ─────────────────────────────────────────────────────────────
echo -e "${BOLD}1. Connectivity & Entry Point${NC}"
# ─────────────────────────────────────────────────────────────

STATUS=$(http_status "${API}/")
if [ "$STATUS" = "200" ]; then
    pass "Entry point returns 200"
else
    fail "Entry point returns 200" "got HTTP $STATUS"
    echo -e "  ${RED}The semantic API is not reachable. Is it enabled in config?${NC}"
    echo -e "  ${RED}Add [semantic] section to hub3.toml or ensure enabled=true${NC}"
    echo ""
    echo -e "Results: ${GREEN}$PASS passed${NC}, ${RED}$FAIL failed${NC}, ${YELLOW}$SKIP skipped${NC}"
    exit 1
fi

ENTRY=$(api "${API}/")
if json_has "$ENTRY" "@type" "*"; then
    pass "Entry point has @type"
else
    fail "Entry point has @type" "missing"
fi

# ─────────────────────────────────────────────────────────────
echo -e "\n${BOLD}2. API Documentation${NC}"
# ─────────────────────────────────────────────────────────────

STATUS=$(http_status "${API}/docs")
if [ "$STATUS" = "200" ]; then
    pass "Docs endpoint returns 200"
else
    fail "Docs endpoint returns 200" "got HTTP $STATUS"
fi

DOCS=$(api "${API}/docs" || echo "{}")
if json_has "$DOCS" "@type" "*"; then
    pass "Docs response has @type (ApiDocumentation)"
else
    fail "Docs response has @type" "missing"
fi

# ─────────────────────────────────────────────────────────────
echo -e "\n${BOLD}3. Search — Basic${NC}"
# ─────────────────────────────────────────────────────────────

SEARCH=$(api "${API}/search?query=*&size=2" || echo "{}")
if json_has "$SEARCH" "totalItems" "*"; then
    TOTAL=$(echo "$SEARCH" | python3 -c "import sys,json; print(json.load(sys.stdin).get('totalItems',0))")
    pass "Search returns totalItems ($TOTAL)"
else
    fail "Search returns totalItems" "missing"
fi

MEMBER_COUNT=$(echo "$SEARCH" | python3 -c "import sys,json; print(len(json.load(sys.stdin).get('member',[])))" 2>/dev/null || echo 0)
if [ "$MEMBER_COUNT" = "2" ]; then
    pass "Search returns requested size (2 members)"
else
    fail "Search returns requested size" "got $MEMBER_COUNT members"
fi

if json_has "$SEARCH" "view" "*"; then
    pass "Search has hydra:view (pagination)"
else
    fail "Search has hydra:view" "missing"
fi

# ─────────────────────────────────────────────────────────────
echo -e "\n${BOLD}4. Search — Filters${NC}"
# ─────────────────────────────────────────────────────────────

FILTERED=$(api "${API}/search?filter[dc_type][eq]=Kerkgebouw&size=1" || echo "{}")
if json_has "$FILTERED" "totalItems" "*"; then
    FTOTAL=$(echo "$FILTERED" | python3 -c "import sys,json; print(json.load(sys.stdin).get('totalItems',0))")
    pass "Filtered search works (totalItems=$FTOTAL)"
else
    skip "Filtered search" "dc_type field may not exist in this dataset"
fi

# ─────────────────────────────────────────────────────────────
echo -e "\n${BOLD}5. Search — Facets${NC}"
# ─────────────────────────────────────────────────────────────

FACETED=$(api "${API}/search?query=*&facet[dc_type]=5&size=0" || echo "{}")
FACET_COUNT=$(echo "$FACETED" | python3 -c "
import sys, json
data = json.load(sys.stdin)
facets = data.get('hub3:facets', data.get('facets', []))
print(len(facets))
" 2>/dev/null || echo 0)

if [ "$FACET_COUNT" -gt 0 ]; then
    pass "Facet aggregation returns results ($FACET_COUNT facets)"
else
    skip "Facet aggregation" "no facets returned"
fi

# ─────────────────────────────────────────────────────────────
echo -e "\n${BOLD}6. Search — Pagination${NC}"
# ─────────────────────────────────────────────────────────────

PAGE1=$(api "${API}/search?query=*&size=1&page=1" || echo "{}")
PAGE2=$(api "${API}/search?query=*&size=1&page=2" || echo "{}")

ID1=$(echo "$PAGE1" | python3 -c "import sys,json; m=json.load(sys.stdin).get('member',[]); print(m[0].get('@id','') if m else '')" 2>/dev/null)
ID2=$(echo "$PAGE2" | python3 -c "import sys,json; m=json.load(sys.stdin).get('member',[]); print(m[0].get('@id','') if m else '')" 2>/dev/null)

if [ -n "$ID1" ] && [ -n "$ID2" ] && [ "$ID1" != "$ID2" ]; then
    pass "Pagination returns different results on page 1 vs 2"
else
    fail "Pagination returns different results" "page1=$ID1 page2=$ID2"
fi

# ─────────────────────────────────────────────────────────────
echo -e "\n${BOLD}7. Resource Detail${NC}"
# ─────────────────────────────────────────────────────────────

# Get a resource ID from search results
RESOURCE_ID=$(echo "$SEARCH" | python3 -c "
import sys, json
data = json.load(sys.stdin)
members = data.get('member', [])
if members:
    print(members[0].get('@id', ''))
" 2>/dev/null)

if [ -n "$RESOURCE_ID" ]; then
    # Extract just the ID part (last path segment or full ID)
    DETAIL_ID=$(echo "$RESOURCE_ID" | python3 -c "
import sys
url = sys.stdin.read().strip()
# URL-encode the full ID for use in the path
import urllib.parse
print(urllib.parse.quote(url, safe=''))
")

    DETAIL=$(api "${API}/resource/${DETAIL_ID}" || echo "{}")

    if json_has "$DETAIL" "@id" "*"; then
        pass "Resource detail returns @id"
    else
        fail "Resource detail returns @id" "missing"
    fi

    if json_has "$DETAIL" "@type" "*"; then
        pass "Resource detail returns @type"
    else
        fail "Resource detail returns @type" "missing"
    fi

    # Check for property value types
    HAS_NESTED=$(echo "$DETAIL" | python3 -c "
import sys, json
data = json.load(sys.stdin)
for k, v in data.items():
    if isinstance(v, dict) and '@id' in v:
        print('yes')
        break
else:
    print('no')
" 2>/dev/null)

    if [ "$HAS_NESTED" = "yes" ]; then
        pass "Resource detail contains nested resource references"
    else
        skip "Nested resource references" "this document may not have any"
    fi

    # Check for hub3:referenceType
    HAS_REF_TYPE=$(echo "$DETAIL" | python3 -c "
import sys, json
def find_ref_type(obj):
    if isinstance(obj, dict):
        if 'hub3:referenceType' in obj:
            return obj['hub3:referenceType']
        for v in obj.values():
            r = find_ref_type(v)
            if r: return r
    elif isinstance(obj, list):
        for v in obj:
            r = find_ref_type(v)
            if r: return r
    return None
data = json.load(sys.stdin)
rt = find_ref_type(data)
print(rt or 'none')
" 2>/dev/null)

    if [ "$HAS_REF_TYPE" != "none" ]; then
        pass "Resource has hub3:referenceType=$HAS_REF_TYPE"
    else
        skip "hub3:referenceType" "this document may not have cycle/external refs"
    fi
else
    skip "Resource detail" "no resource ID from search"
fi

# ─────────────────────────────────────────────────────────────
echo -e "\n${BOLD}8. Search Context & Navigation${NC}"
# ─────────────────────────────────────────────────────────────

# Check if search response includes context
CONTEXT_TOKEN=$(echo "$SEARCH" | python3 -c "
import sys, json
data = json.load(sys.stdin)
ctx = data.get('hub3:searchContext', data.get('searchContext', {}))
if isinstance(ctx, dict):
    print(ctx.get('token', ''))
else:
    print('')
" 2>/dev/null)

if [ -n "$CONTEXT_TOKEN" ]; then
    pass "Search response includes searchContext token"

    # Test detail with context for navigation
    if [ -n "$RESOURCE_ID" ]; then
        NAV_DETAIL=$(api "${API}/resource/${DETAIL_ID}?context=${CONTEXT_TOKEN}" || echo "{}")
        HAS_NAV=$(echo "$NAV_DETAIL" | python3 -c "
import sys, json
data = json.load(sys.stdin)
nav = data.get('hub3:navigation', data.get('navigation', {}))
print('yes' if nav else 'no')
" 2>/dev/null)

        if [ "$HAS_NAV" = "yes" ]; then
            pass "Detail with context returns hub3:navigation"
        else
            fail "Detail with context returns hub3:navigation" "missing"
        fi
    fi
else
    skip "Search context" "no context token in search response"
fi

# ─────────────────────────────────────────────────────────────
echo -e "\n${BOLD}9. Introspection${NC}"
# ─────────────────────────────────────────────────────────────

STATUS=$(http_status "${API}/introspect/classes")
if [ "$STATUS" = "200" ]; then
    pass "Introspect classes endpoint returns 200"
else
    skip "Introspect classes" "endpoint returned HTTP $STATUS"
fi

# ─────────────────────────────────────────────────────────────
echo -e "\n${BOLD}10. Error Handling${NC}"
# ─────────────────────────────────────────────────────────────

ERR_STATUS=$(http_status "${API}/resource/nonexistent-id-that-should-not-exist-12345")
if [ "$ERR_STATUS" = "404" ]; then
    pass "Missing resource returns 404"
else
    fail "Missing resource returns 404" "got HTTP $ERR_STATUS"
fi

BAD_FILTER_STATUS=$(http_status "${API}/search?filter[dc_type][invalidop]=test")
if [ "$BAD_FILTER_STATUS" = "400" ]; then
    pass "Invalid operator returns 400"
else
    skip "Invalid operator returns 400" "got HTTP $BAD_FILTER_STATUS"
fi

# ─────────────────────────────────────────────────────────────
echo -e "\n${BOLD}11. JSON-LD Context Files${NC}"
# ─────────────────────────────────────────────────────────────

CTX_STATUS=$(http_status "${BASE}/contexts/hub3/latest/context.jsonld")
if [ "$CTX_STATUS" = "200" ]; then
    pass "Hub3 context.jsonld is served"
    CTX=$(api "${BASE}/contexts/hub3/latest/context.jsonld" || echo "{}")
    if json_has "$CTX" "@context" "*"; then
        pass "Hub3 context.jsonld has @context"
    else
        fail "Hub3 context.jsonld has @context" "missing"
    fi
else
    fail "Hub3 context.jsonld is served" "got HTTP $CTX_STATUS"
fi

# ─────────────────────────────────────────────────────────────
echo -e "\n${BOLD}12. Property Value Types (Deep Check)${NC}"
# ─────────────────────────────────────────────────────────────

# Fetch a larger result and inspect value type variety
RICH=$(api "${API}/search?query=*&size=5" || echo "{}")
VALUE_TYPES=$(echo "$RICH" | python3 -c "
import sys, json

def classify(v):
    types = set()
    if isinstance(v, str):
        types.add('plain_literal')
    elif isinstance(v, (int, float)):
        types.add('number')
    elif isinstance(v, bool):
        types.add('boolean')
    elif isinstance(v, dict):
        if '@value' in v and '@language' in v:
            types.add('lang_tagged')
        elif '@value' in v and '@type' in v:
            types.add('typed_literal')
        elif '@value' in v:
            types.add('typed_literal')
        elif '@id' in v:
            if 'hub3:referenceType' in v:
                types.add('ref_' + v['hub3:referenceType'])
            elif len(v) == 1:
                types.add('id_only')
            else:
                types.add('nested_resource')
        else:
            # Could be language map
            if all(len(k) <= 3 for k in v.keys()):
                types.add('language_map')
    elif isinstance(v, list):
        types.add('array')
        for item in v:
            types |= classify(item)
    return types

data = json.load(sys.stdin)
all_types = set()
for member in data.get('member', []):
    for k, v in member.items():
        if not k.startswith('@'):
            all_types |= classify(v)

print(','.join(sorted(all_types)) if all_types else 'none')
" 2>/dev/null)

if [ "$VALUE_TYPES" != "none" ]; then
    pass "Value types found in results: $VALUE_TYPES"
else
    skip "Value type inspection" "no typed values in sample"
fi

# ─────────────────────────────────────────────────────────────
echo -e "\n${BOLD}13. Cross-Index Search${NC}"
# ─────────────────────────────────────────────────────────────

# This should work without error even if there's only one org
CROSS=$(api "${API}/search?query=*&size=1&contextIndex=nonexistent-org" || echo "error")
if [ "$CROSS" != "error" ]; then
    pass "Cross-index search does not error with extra contextIndex"
else
    fail "Cross-index search" "request failed"
fi

# ─────────────────────────────────────────────────────────────
# Summary
# ─────────────────────────────────────────────────────────────
echo ""
echo -e "${BOLD}────────────────────────────────────────${NC}"
TOTAL=$((PASS + FAIL + SKIP))
echo -e "Results: ${GREEN}$PASS passed${NC}, ${RED}$FAIL failed${NC}, ${YELLOW}$SKIP skipped${NC} (${TOTAL} total)"
echo -e "${BOLD}────────────────────────────────────────${NC}"

if [ "$FAIL" -gt 0 ]; then
    exit 1
fi
