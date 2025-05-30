#!/bin/bash
# PWA Icon Generation Script
# Usage: ./generate-pwa-icons.sh

set -e

# Ensure we're in the correct directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
STATIC_DIR="$SCRIPT_DIR/app/static"

echo "🎨 PWA Icon Generator"
echo "=================="

# Navigate to static directory
if [ ! -d "$STATIC_DIR" ]; then
    echo "❌ app/static directory not found"
    echo "   Please run this script from the project root"
    exit 1
fi

cd "$STATIC_DIR"

# Check if source logo exists
if [ ! -f "logo.png" ]; then
    echo "❌ logo.png not found in app/static/"
    echo "   Please ensure logo.png exists in app/static/"
    exit 1
fi

# Check if ImageMagick is installed
if ! command -v magick &> /dev/null; then
    echo "❌ ImageMagick not found"
    echo "   Please install with: brew install imagemagick"
    echo "   Or use the online tool method described in PWA_ICON_GENERATION_GUIDE.md"
    exit 1
fi

echo "📂 Working directory: $STATIC_DIR"
echo "🖼️  Source logo: logo.png"

# Get source logo dimensions for validation
LOGO_INFO=$(magick identify logo.png)
echo "📏 Source logo info: $LOGO_INFO"

# Create icons directory
mkdir -p icons

# Define icon sizes for PWA
PWA_SIZES=(48 72 96 128 144 152 192 384 512)
FAVICON_SIZES=(16 32)
EXTRA_SIZES=(1024)

echo ""
echo "🔄 Generating PWA icons..."

# Generate PWA icons
for size in "${PWA_SIZES[@]}"; do
    echo "  📱 ${size}x${size}..."
    magick logo.png -resize ${size}x${size} -filter Lanczos icons/icon-${size}x${size}.png
done

echo ""
echo "🌐 Generating favicons..."

# Generate favicons
for size in "${FAVICON_SIZES[@]}"; do
    echo "  🔗 ${size}x${size}..."
    magick logo.png -resize ${size}x${size} -unsharp 0x1+1.0+0.05 favicon-${size}x${size}.png
done

# Generate special favicon files
echo "  📄 favicon.ico..."
magick logo.png -resize 32x32 favicon.ico

echo "  🔄 Updating main favicon.png..."
magick logo.png -resize 32x32 -unsharp 0x1+1.0+0.05 favicon.png

echo ""
echo "🍎 Generating Apple touch icon..."
magick logo.png -resize 180x180 -filter Lanczos apple-touch-icon.png

echo ""
echo "🏪 Generating extra sizes..."
for size in "${EXTRA_SIZES[@]}"; do
    echo "  📦 ${size}x${size}..."
    magick logo.png -resize ${size}x${size} -filter Lanczos icons/icon-${size}x${size}.png
done

echo ""
echo "✅ PWA icons generated successfully!"
echo ""
echo "📋 Generated files:"
echo "==================="

# List generated files
echo "Favicons:"
ls -la favicon* 2>/dev/null || echo "  No favicon files found"

echo ""
echo "Apple touch icon:"
ls -la apple-touch-icon.png 2>/dev/null || echo "  apple-touch-icon.png not found"

echo ""
echo "PWA icons:"
ls -la icons/ 2>/dev/null || echo "  No icons directory found"

echo ""
echo "📊 File sizes:"
du -h favicon* apple-touch-icon.png icons/*.png 2>/dev/null | sort -k1 -h

echo ""
echo "🎯 Next steps:"
echo "1. Review the generated icons"
echo "2. Update app.html with favicon links"
echo "3. Create/update manifest.json"
echo "4. Enable PWA in vite.config.ts"
echo "5. Test PWA functionality"
echo ""
echo "📖 See PWA_ICON_GENERATION_GUIDE.md for detailed instructions" 