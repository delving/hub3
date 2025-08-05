# API Documentation Build System

This build system generates professional PDF documentation from the Hub3 API markdown files with support for custom branding.

## Prerequisites

- **pandoc** (2.0 or higher)
- **XeLaTeX** or **pdfLaTeX** (usually part of TeX Live or MiKTeX)
- **Git** (for including commit SHA)

### Installing Dependencies

#### Ubuntu/Debian
```bash
sudo apt-get update
sudo apt-get install pandoc texlive-xetex texlive-fonts-recommended texlive-latex-extra
```

#### macOS
```bash
brew install pandoc
brew install --cask mactex  # Or basictex for smaller installation
```

#### Windows
- Install [Pandoc](https://pandoc.org/installing.html)
- Install [MiKTeX](https://miktex.org/download)

## Quick Start

### Build with Default Delving Branding
```bash
./build-api-docs.sh
```

This creates PDFs for both V1 and V2 API documentation in the `pdf/` directory.

### Build Only One Version
```bash
# Build only V1 documentation
./build-api-docs.sh --v1-only

# Build only V2 documentation  
./build-api-docs.sh --v2-only
```

### Build with Custom Branding
```bash
./build-api-docs.sh --brand customer-name
```

## Output Files

PDFs are generated in the `pdf/` directory with the following naming convention:
```
{API_VERSION}_API_DOCUMENTATION_{BRAND}_{TIMESTAMP}_{GIT_SHA}.pdf
```

Additionally, symbolic links to the latest versions are created:
```
V1_API_DOCUMENTATION_{BRAND}_latest.pdf
V2_API_DOCUMENTATION_{BRAND}_latest.pdf
```

### Example Output
```
pdf/
├── V1_API_DOCUMENTATION_delving_20250727_143022_a3b5c7d.pdf
├── V1_API_DOCUMENTATION_delving_latest.pdf -> V1_API_DOCUMENTATION_delving_20250727_143022_a3b5c7d.pdf
├── V2_API_DOCUMENTATION_delving_20250727_143022_a3b5c7d.pdf
└── V2_API_DOCUMENTATION_delving_latest.pdf -> V2_API_DOCUMENTATION_delving_20250727_143022_a3b5c7d.pdf
```

## Custom Branding

### Setting Up a New Brand

1. Create a new brand directory:
   ```bash
   mkdir -p doc-templates/brands/your-brand
   ```

2. Create a configuration file `doc-templates/brands/your-brand/config.tex`:
   ```latex
   % Your Brand Configuration
   \definecolor{brandprimary}{RGB}{100, 50, 200}    % Your primary color
   \definecolor{brandsecondary}{RGB}{50, 25, 100}   % Your secondary color
   
   \def\brandname{Your Organization Name}
   \def\brandwebsite{https://www.your-website.com}
   ```

3. Add your logo as `doc-templates/brands/your-brand/logo.png` (or `.eps`)

4. Build with your brand:
   ```bash
   ./build-api-docs.sh --brand your-brand
   ```

### Brand Directory Structure
```
doc-templates/
├── api-template.tex          # Main LaTeX template
└── brands/
    ├── delving/             # Default Delving brand
    │   ├── config.tex       # Brand configuration
    │   └── logo.png         # Brand logo
    ├── example-customer/    # Example configuration
    │   └── config.tex
    └── README.md           # Branding guide
```

## Features

### Automatic Metadata
Each PDF includes:
- **Build timestamp**: When the PDF was generated
- **Git SHA**: Short commit hash for traceability
- **API version**: V1 or V2
- **Generation date**: Current date

### Professional Formatting
- Table of contents with clickable links
- Syntax highlighting for code examples
- Branded headers and footers
- Consistent typography and spacing
- Responsive tables and lists

### Version Control Integration
The build system automatically:
- Detects the current Git commit SHA
- Includes it in the PDF metadata
- Falls back gracefully if not in a Git repository

## Customization

### Modifying the Template
Edit `doc-templates/api-template.tex` to change:
- Page layout and margins
- Font choices
- Header/footer design
- Color schemes
- Section formatting

### Adding Sections to Documentation
Edit the source markdown files:
- `V1_API_DOCUMENTATION.md`
- `V2_API_DOCUMENTATION.md`

Then rebuild the PDFs.

## Troubleshooting

### Common Issues

1. **"Missing dependencies" error**
   - Install pandoc and XeLaTeX (see Prerequisites)

2. **Logo not appearing**
   - Check file exists: `doc-templates/brands/{brand}/logo.png`
   - Ensure correct file format (PNG or EPS)

3. **Build fails with LaTeX errors**
   - Check for special characters in markdown that need escaping
   - Verify custom brand config.tex syntax

4. **Git SHA shows "no-git"**
   - Ensure you're in a Git repository
   - Commit your changes before building

### Debug Mode
For detailed output, run pandoc directly:
```bash
pandoc V1_API_DOCUMENTATION.md \
  --template=doc-templates/api-template.tex \
  --pdf-engine=xelatex \
  -V BRANDCONFIG="doc-templates/brands/delving/config.tex" \
  -V apiversion="v1" \
  --verbose \
  -o test.pdf
```

## Advanced Usage

### Batch Building for Multiple Brands
```bash
#!/bin/bash
for brand in delving customer1 customer2; do
  ./build-api-docs.sh --brand $brand
done
```

### Continuous Integration
Add to your CI pipeline:
```yaml
- name: Build API Documentation
  run: |
    ./build-api-docs.sh --brand ${{ env.BRAND_NAME }}
    
- name: Upload PDFs
  uses: actions/upload-artifact@v3
  with:
    name: api-documentation
    path: pdf/*_latest.pdf
```

## Contributing

To improve the build system:
1. Test changes with multiple brands
2. Ensure backwards compatibility
3. Update this documentation
4. Submit a pull request

## License

This build system is part of Hub3 and follows the same license terms.