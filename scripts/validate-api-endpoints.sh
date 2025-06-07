#!/bin/bash

# validate-api-endpoints.sh
# CI script to validate that all API endpoints called from frontend exist in the backend

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}🔍 Validating API endpoint consistency between frontend and backend...${NC}"

# Extract unique API endpoints from frontend code
echo "Extracting API endpoints from frontend..."

# Extract API calls from TypeScript/JavaScript/Svelte files
FRONTEND_ENDPOINTS=$(find app/src -name "*.ts" -o -name "*.js" -o -name "*.svelte" | \
    xargs grep -ho "'/api/[^'\"]*'" | \
    sed "s/'//g" | \
    sort | uniq)

# Also check for template strings and variables
TEMPLATE_ENDPOINTS=$(find app/src -name "*.ts" -o -name "*.js" -o -name "*.svelte" | \
    xargs grep -ho "\`/api/[^\`]*\`" | \
    sed "s/\`//g" | \
    sed 's/\${[^}]*}/\{id\}/g' | \
    sort | uniq)

# Extract from concatenated strings like '/api/admin/bookings' + id
CONCAT_ENDPOINTS=$(find app/src -name "*.ts" -o -name "*.js" -o -name "*.svelte" | \
    xargs grep -ho '"/api/[^"]*"' | \
    sed 's/"//g' | \
    sort | uniq)

# Combine and deduplicate, filter out incomplete patterns
ALL_FRONTEND_ENDPOINTS=$(echo -e "$FRONTEND_ENDPOINTS\n$TEMPLATE_ENDPOINTS\n$CONCAT_ENDPOINTS" | \
    sort | uniq | \
    grep -v '^$' | \
    grep -v '\${' | \
    grep -v '?' | \
    grep -E '^/api/[a-zA-Z0-9/_-]+(\{[^}]+\})?$')

echo "Found $(echo "$ALL_FRONTEND_ENDPOINTS" | wc -l) unique API endpoints in frontend code:"
echo "$ALL_FRONTEND_ENDPOINTS" | sed 's/^/  /'

# Extract backend routes from main.go
echo ""
echo "Extracting registered routes from backend..."

# Extract routes from public API group (with /api prefix)
PUBLIC_ROUTES=$(grep -A 20 'publicAPI := fuego.Group(s, apiPrefix)' cmd/server/main.go | \
    grep -E 'fuego\.(Get|Post|Put|Delete)' | \
    grep -o '"/[^"]*"' | \
    sed 's/"//g' | \
    sed 's/{[^}]*}/{id}/g' | \
    sed 's|^|/api|')

# Extract routes from protected group (with /api prefix)  
PROTECTED_ROUTES=$(grep -A 30 'protected := fuego.Group(s, apiPrefix)' cmd/server/main.go | \
    grep -E 'fuego\.(Get|Post|Put|Delete)' | \
    grep -o '"/[^"]*"' | \
    sed 's/"//g' | \
    sed 's/{[^}]*}/{id}/g' | \
    sed 's|^|/api|')

# Extract routes from admin group (with /api/admin prefix)
ADMIN_ROUTES=$(sed -n '/admin := fuego.Group(s, apiPrefix+"\/admin")/,/fuego.GetStd(s, "\/swagger"/p' cmd/server/main.go | \
    grep -E 'fuego\.(Get|Post|Put|Delete)' | \
    grep -o '"/[^"]*"' | \
    sed 's/"//g' | \
    sed 's/{[^}]*}/{id}/g' | \
    sed 's|^|/api/admin|')

# Combine all routes
BACKEND_ROUTES=$(echo -e "$PUBLIC_ROUTES\n$PROTECTED_ROUTES\n$ADMIN_ROUTES" | sort | uniq | grep -v '^$')

echo "Found $(echo "$BACKEND_ROUTES" | wc -l) registered routes in backend:"
echo "$BACKEND_ROUTES" | sed 's/^/  /'

# Check for missing endpoints
echo ""
echo -e "${YELLOW}🔎 Checking for frontend endpoints missing in backend...${NC}"

MISSING_ENDPOINTS=""
FOUND_ENDPOINTS=""

while IFS= read -r endpoint; do
    if [ -n "$endpoint" ]; then
        # Convert frontend endpoint to backend route pattern
        BACKEND_PATTERN=$(echo "$endpoint" | sed 's/{[^}]*}/{id}/g')
        
        if echo "$BACKEND_ROUTES" | grep -q "^$BACKEND_PATTERN$"; then
            FOUND_ENDPOINTS="$FOUND_ENDPOINTS\n  ✅ $endpoint"
        else
            MISSING_ENDPOINTS="$MISSING_ENDPOINTS\n  ❌ $endpoint"
        fi
    fi
done <<< "$ALL_FRONTEND_ENDPOINTS"

# Also check swagger documentation if available
SWAGGER_PATHS=""
if [ -f "openapi.json" ]; then
    echo ""
    echo "Cross-referencing with OpenAPI specification..."
    SWAGGER_PATHS=$(jq -r '.paths | keys[]' openapi.json 2>/dev/null || echo "")
fi

# Report results
echo ""
if [ -n "$MISSING_ENDPOINTS" ]; then
    echo -e "${RED}❌ VALIDATION FAILED: Found API endpoints in frontend that are missing in backend:${NC}"
    echo -e "$MISSING_ENDPOINTS"
    
    if [ -n "$SWAGGER_PATHS" ]; then
        echo ""
        echo -e "${YELLOW}📋 Available endpoints in OpenAPI spec:${NC}"
        echo "$SWAGGER_PATHS" | sed 's/^/  /'
    fi
    
    echo ""
    echo -e "${YELLOW}💡 Action needed:${NC}"
    echo "  1. Implement missing backend endpoints"
    echo "  2. Or remove unused frontend API calls"
    echo "  3. Or update frontend to use correct endpoint paths"
    
    exit 1
else
    echo -e "${GREEN}✅ VALIDATION PASSED: All frontend API endpoints are implemented in backend${NC}"
fi

if [ -n "$FOUND_ENDPOINTS" ]; then
    echo ""
    echo -e "${GREEN}✅ Validated endpoints:${NC}"
    echo -e "$FOUND_ENDPOINTS"
fi

echo ""
echo -e "${GREEN}🎉 API endpoint validation completed successfully!${NC}" 