# API Documentation Branding Guide

This directory contains brand configurations for generating customized API documentation PDFs.

## Directory Structure

```
brands/
├── delving/          # Default Delving branding
│   ├── config.tex    # Brand colors and settings
│   └── logo.png      # Brand logo (or logo.eps)
├── example-customer/ # Example customer brand
│   ├── config.tex    
│   └── logo.png      
└── README.md         # This file
```

## Creating a New Brand

1. Copy the `example-customer` directory:
   ```bash
   cp -r example-customer your-brand-name
   ```

2. Edit `your-brand-name/config.tex` to customize:
   - Brand colors (RGB values)
   - Organization name
   - Website URL
   - Any additional LaTeX settings

3. Add your logo as either:
   - `logo.png` (recommended, transparent background)
   - `logo.eps` (for better quality at any size)

4. Build documentation with your brand:
   ```bash
   ./build-api-docs.sh --brand your-brand-name
   ```

## Brand Configuration Options

### Required Settings in config.tex

```latex
% Brand colors
\definecolor{brandprimary}{RGB}{R, G, B}      % Primary brand color
\definecolor{brandsecondary}{RGB}{R, G, B}    % Secondary color (headings)

% Brand information
\def\brandname{Your Organization Name}
\def\brandwebsite{https://www.yoursite.com}
```

### Optional Settings

```latex
% Additional colors
\definecolor{brandaccent}{RGB}{R, G, B}       % Accent color
\definecolor{brandwarning}{RGB}{R, G, B}      % Warning/alert color

% Additional text
\def\brandtagline{Your tagline or motto}
\def\branddepartment{IT Department}
\def\brandcontact{api-support@yoursite.com}
```

## Logo Requirements

- **Format**: PNG (with transparency) or EPS
- **Recommended size**: At least 300x100 pixels
- **Aspect ratio**: Will be scaled to 1.5cm height
- **File name**: Must be `logo.png` or `logo.eps`

## Color Selection Tips

1. Use your organization's official brand colors
2. Ensure sufficient contrast for readability
3. Test print in both color and grayscale
4. Consider accessibility (WCAG guidelines)

## Example Brands

### Delving (Default)
- Primary: RGB(41, 118, 176) - Blue
- Secondary: RGB(25, 85, 140) - Dark Blue

### Example Customer
- Primary: RGB(0, 120, 215) - Bright Blue
- Secondary: RGB(0, 78, 170) - Dark Blue

## Testing Your Brand

After creating your brand configuration:

1. Build a test PDF:
   ```bash
   ./build-api-docs.sh --brand your-brand-name --v1-only
   ```

2. Check the output in `pdf/` directory

3. Verify:
   - Logo appears correctly
   - Colors are as expected
   - Text is readable
   - Links are visible

## Troubleshooting

- **Logo not appearing**: Check file exists and path is correct
- **Colors not changing**: Ensure config.tex syntax is valid
- **Build fails**: Check LaTeX error messages, usually syntax issues

For help, consult the main build script documentation or LaTeX documentation.