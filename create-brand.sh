#!/bin/bash
# Helper script to create a new brand configuration
# Usage: ./create-brand.sh BRAND_NAME

set -e

BRAND_NAME="$1"
TEMPLATES_DIR="doc-templates/brands"

if [ -z "$BRAND_NAME" ]; then
  echo "Usage: $0 BRAND_NAME"
  echo "Example: $0 rijksmuseum"
  exit 1
fi

# Sanitize brand name (lowercase, alphanumeric and hyphens only)
BRAND_DIR=$(echo "$BRAND_NAME" | tr '[:upper:]' '[:lower:]' | sed 's/[^a-z0-9-]/-/g')
BRAND_PATH="$TEMPLATES_DIR/$BRAND_DIR"

if [ -d "$BRAND_PATH" ]; then
  echo "Error: Brand '$BRAND_DIR' already exists at $BRAND_PATH"
  exit 1
fi

echo "Creating new brand: $BRAND_DIR"
mkdir -p "$BRAND_PATH"

# Create config.tex with placeholders
cat > "$BRAND_PATH/config.tex" <<EOF
% $BRAND_NAME Brand Configuration
% Generated on $(date)

% Brand colors - CUSTOMIZE THESE
\definecolor{brandprimary}{RGB}{0, 100, 200}      % Primary brand color
\definecolor{brandsecondary}{RGB}{0, 50, 100}     % Secondary color for headings
\definecolor{lightgray}{RGB}{240, 240, 240}       % Light gray for backgrounds
\definecolor{codebg}{RGB}{248, 248, 248}          % Code background

% Optional accent colors
\definecolor{brandaccent}{RGB}{255, 152, 0}       % Accent color
\definecolor{brandsuccess}{RGB}{76, 175, 80}      % Success messages
\definecolor{brandwarning}{RGB}{255, 152, 0}      % Warning messages
\definecolor{branderror}{RGB}{244, 67, 54}        % Error messages

% Brand information - CUSTOMIZE THESE
\def\brandname{$BRAND_NAME}
\def\brandwebsite{https://www.example.com}
\def\brandcontact{api@example.com}

% Optional: Company details (for footer)
\def\companyaddress{%
  Your Street 123\\
  1234 AB City\\
  Country
}

\def\companycontact{%
  Tel: +31 (0) 12 345 6789\\
  Email: \brandcontact
}
EOF

# Create a README for the brand
cat > "$BRAND_PATH/README.md" <<EOF
# $BRAND_NAME Brand Configuration

This directory contains the branding configuration for $BRAND_NAME.

## Setup Instructions

1. **Add your logo**:
   - Save your logo as \`logo.png\` (recommended) or \`logo.eps\` in this directory
   - Recommended dimensions: at least 300x100 pixels
   - Use transparent background for best results

2. **Customize colors** in \`config.tex\`:
   - Update the RGB values for \`brandprimary\` and \`brandsecondary\`
   - These should match your organization's brand guidelines

3. **Update organization info** in \`config.tex\`:
   - Set the correct \`brandname\`
   - Update \`brandwebsite\` 
   - Set \`brandcontact\` email
   - Optionally update company address and contact details

4. **Test your brand**:
   \`\`\`bash
   ../../build-api-docs.sh --brand $BRAND_DIR --v1-only
   \`\`\`

## Files

- \`config.tex\` - LaTeX configuration with colors and brand information
- \`logo.png\` or \`logo.eps\` - Your brand logo (you need to add this)
- \`README.md\` - This file

## Color Guidelines

Current placeholder colors:
- Primary: RGB(0, 100, 200) - Blue
- Secondary: RGB(0, 50, 100) - Dark Blue

Replace these with your brand colors.
EOF

echo "✓ Brand directory created at: $BRAND_PATH"
echo ""
echo "Next steps:"
echo "1. Add your logo to: $BRAND_PATH/logo.png"
echo "2. Edit brand colors in: $BRAND_PATH/config.tex"
echo "3. Update organization info in: $BRAND_PATH/config.tex"
echo "4. Test with: ./build-api-docs.sh --brand $BRAND_DIR"
echo ""
echo "See $BRAND_PATH/README.md for detailed instructions."