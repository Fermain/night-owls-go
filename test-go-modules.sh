#!/bin/bash
# Test script to identify which Go module is causing the hang

echo "Testing Go module downloads..."
echo "Go version: $(go version)"
echo "GOPROXY: ${GOPROXY:-not set}"
echo "---"

# Create a temporary directory
TEMP_DIR=$(mktemp -d)
cd "$TEMP_DIR"

# Copy go.mod and go.sum
cp /Users/oliver/Repos/night-owls-go/go.mod .
cp /Users/oliver/Repos/night-owls-go/go.sum .

# Try downloading with verbose output
echo "Attempting go mod download with verbose output..."
timeout 30s go mod download -x 2>&1 | tee download.log

if [ $? -eq 124 ]; then
    echo "Download timed out after 30 seconds"
    echo "Last few lines of output:"
    tail -20 download.log
    
    echo "---"
    echo "Trying to identify problematic module..."
    
    # Extract module names from go.mod
    grep -E "^\s*[^/]+\.[^/]+/" go.mod | while read -r line; do
        module=$(echo "$line" | awk '{print $1}')
        echo "Testing module: $module"
        timeout 10s go get -d "$module" 2>&1
        if [ $? -eq 124 ]; then
            echo "  ⚠️  Module $module timed out!"
        else
            echo "  ✓ Module $module downloaded successfully"
        fi
    done
else
    echo "Download completed successfully!"
fi

# Cleanup
cd -
rm -rf "$TEMP_DIR" 