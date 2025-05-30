# PWA Icon Generation Guide

## Overview

This guide provides a complete procedure for generating all necessary PWA icons and favicons from your source logo (`app/static/logo.png`).

## Prerequisites

- Source logo: `app/static/logo.png` (should be high resolution, preferably 1024x1024px or larger)
- ImageMagick installed (`brew install imagemagick` on macOS)
- Or use online tools like [Favicon Generator](https://realfavicongenerator.net/)

## Icon Specifications

### Required PWA Icons

| Size | Purpose | Filename |
|------|---------|----------|
| 16x16 | Browser favicon | `favicon-16x16.png` |
| 32x32 | Browser favicon | `favicon-32x32.png` |
| 48x48 | PWA icon | `icon-48x48.png` |
| 72x72 | PWA icon | `icon-72x72.png` |
| 96x96 | PWA icon | `icon-96x96.png` |
| 128x128 | PWA icon | `icon-128x128.png` |
| 144x144 | PWA icon | `icon-144x144.png` |
| 152x152 | PWA icon | `icon-152x152.png` |
| 192x192 | PWA icon | `icon-192x192.png` |
| 384x384 | PWA icon | `icon-384x384.png` |
| 512x512 | PWA icon | `icon-512x512.png` |
| 180x180 | Apple touch icon | `apple-touch-icon.png` |

### Optional but Recommended

| Size | Purpose | Filename |
|------|---------|----------|
| 1024x1024 | App store | `icon-1024x1024.png` |
| 32x32 | ICO favicon | `favicon.ico` |

## Automated Generation Script

### Method 1: Using ImageMagick (Recommended)

Create a script to generate all icons:

```bash
#!/bin/bash
# Save as: generate-pwa-icons.sh
# Usage: ./generate-pwa-icons.sh

# Ensure we're in the app/static directory
cd "$(dirname "$0")/app/static" || exit 1

# Check if source logo exists
if [ ! -f "logo.png" ]; then
    echo "❌ logo.png not found in app/static/"
    exit 1
fi

echo "🎨 Generating PWA icons from logo.png..."

# Create icons directory
mkdir -p icons

# Define icon sizes
SIZES=(16 32 48 72 96 128 144 152 192 384 512 1024)

# Generate standard PWA icons
for size in "${SIZES[@]}"; do
    echo "📱 Generating ${size}x${size} icon..."
    magick logo.png -resize ${size}x${size} icons/icon-${size}x${size}.png
done

# Generate specific favicons
echo "🌐 Generating favicons..."
magick logo.png -resize 16x16 favicon-16x16.png
magick logo.png -resize 32x32 favicon-32x32.png
magick logo.png -resize 32x32 favicon.ico

# Generate Apple touch icon
echo "🍎 Generating Apple touch icon..."
magick logo.png -resize 180x180 apple-touch-icon.png

# Update main favicon (keep existing filename for compatibility)
echo "🔄 Updating main favicon..."
magick logo.png -resize 32x32 favicon.png

echo "✅ PWA icons generated successfully!"
echo ""
echo "📋 Generated files:"
ls -la favicon* apple-touch-icon.png icons/ 2>/dev/null || echo "Some files may not have been generated"
```

### Method 2: Using Online Tools

1. **Visit [RealFaviconGenerator](https://realfavicongenerator.net/)**
2. **Upload your `logo.png`**
3. **Configure each platform:**
   - **Desktop browsers**: Enable favicon, adjust if needed
   - **iOS**: Enable Apple touch icon
   - **Android Chrome**: Enable PWA icons
   - **Windows Metro**: Configure tile colors
4. **Download the generated package**
5. **Extract to `app/static/`**

## Manual Process (ImageMagick)

If you prefer manual control:

```bash
# Navigate to the static directory
cd app/static

# Standard PWA icons
magick logo.png -resize 192x192 icons/icon-192x192.png
magick logo.png -resize 512x512 icons/icon-512x512.png
magick logo.png -resize 384x384 icons/icon-384x384.png
magick logo.png -resize 144x144 icons/icon-144x144.png
magick logo.png -resize 128x128 icons/icon-128x128.png
magick logo.png -resize 96x96 icons/icon-96x96.png
magick logo.png -resize 72x72 icons/icon-72x72.png
magick logo.png -resize 48x48 icons/icon-48x48.png

# Favicons
magick logo.png -resize 32x32 favicon.png
magick logo.png -resize 32x32 favicon.ico
magick logo.png -resize 16x16 favicon-16x16.png

# Apple touch icon
magick logo.png -resize 180x180 apple-touch-icon.png
```

## Quality Optimization

### For Better Results:
1. **Source logo should be:**
   - Minimum 1024x1024px
   - PNG format with transparency
   - Square aspect ratio
   - High contrast for small sizes

2. **ImageMagick optimization flags:**
   ```bash
   # For better small icon quality
   magick logo.png -resize 32x32 -unsharp 0x1+1.0+0.05 favicon.png
   
   # For crisp edges
   magick logo.png -resize 192x192 -filter Lanczos icons/icon-192x192.png
   ```

## Directory Structure

After generation, your `app/static/` should contain:

```
app/static/
├── logo.png                 # Source logo
├── favicon.png              # 32x32 favicon
├── favicon.ico              # ICO format favicon
├── favicon-16x16.png        # 16x16 favicon
├── favicon-32x32.png        # 32x32 favicon
├── apple-touch-icon.png     # 180x180 Apple icon
├── icons/
│   ├── icon-48x48.png
│   ├── icon-72x72.png
│   ├── icon-96x96.png
│   ├── icon-128x128.png
│   ├── icon-144x144.png
│   ├── icon-152x152.png
│   ├── icon-192x192.png
│   ├── icon-384x384.png
│   ├── icon-512x512.png
│   └── icon-1024x1024.png
├── manifest.json            # PWA manifest
└── sw.js                    # Service worker
```

## Verification

### Check Generated Icons:
```bash
# Verify all icons were generated
find app/static -name "*.png" -o -name "*.ico" | sort

# Check icon sizes
file app/static/icons/*.png
```

### Test in Browser:
1. **Favicon**: Check browser tab icon
2. **PWA**: Use browser dev tools → Application → Manifest
3. **Apple**: Test "Add to Home Screen" on iOS Safari

## Integration with Manifest

After generating icons, update your `manifest.json`:

```json
{
  "name": "Mount Moreland Night Owls",
  "short_name": "Night Owls",
  "icons": [
    {
      "src": "/icons/icon-48x48.png",
      "sizes": "48x48",
      "type": "image/png",
      "purpose": "any"
    },
    {
      "src": "/icons/icon-192x192.png",
      "sizes": "192x192",
      "type": "image/png",
      "purpose": "any maskable"
    },
    {
      "src": "/icons/icon-512x512.png",
      "sizes": "512x512",
      "type": "image/png",
      "purpose": "any maskable"
    }
  ]
}
```

## Automation in Build Process

Add to your `package.json` scripts:

```json
{
  "scripts": {
    "generate-icons": "./generate-pwa-icons.sh",
    "prebuild": "npm run generate-icons"
  }
}
```

## Troubleshooting

### Common Issues:

1. **Icons appear blurry**: Use higher resolution source logo
2. **ImageMagick not found**: Install with `brew install imagemagick`
3. **Manifest errors**: Validate at [PWA Manifest Validator](https://manifest-validator.appspot.com/)
4. **Icons not updating**: Clear browser cache or use hard refresh

### Quality Tips:

1. **For logos with text**: Ensure text remains readable at 48x48px
2. **For detailed logos**: Consider simplified version for small sizes
3. **Color consistency**: Test on light and dark backgrounds
4. **Maskable icons**: Ensure important elements are within safe zone

## Next Steps

After generating icons:

1. **Update `app.html`** with favicon links
2. **Create/update `manifest.json`**
3. **Enable PWA in `vite.config.ts`**
4. **Test PWA installation**
5. **Deploy and verify on actual devices** 