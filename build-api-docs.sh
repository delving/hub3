#!/bin/bash
# Build API Documentation PDFs with Branding Support
# Usage: ./build-api-docs.sh [--brand BRAND_NAME] [--v1-only | --v2-only]
# Output will be in pdf/ directory with timestamp and git sha

set -e

# Default values
BRAND="delving"
BUILD_V1=true
BUILD_V2=true
PDF_DIR="pdf"
# For development/testing, you can use: PDF_DIR="tmp/pdf"
TEMPLATES_DIR="doc-templates"
TIMESTAMP=$(date +"%Y%m%d-%H%M%S")
GIT_SHA=$(git rev-parse --short HEAD 2>/dev/null || echo "no-git")

# Parse command line arguments
while [[ $# -gt 0 ]]; do
  key="$1"
  case $key in
    --brand)
      BRAND="$2"
      shift
      shift
      ;;
    --v1-only)
      BUILD_V1=true
      BUILD_V2=false
      shift
      ;;
    --v2-only)
      BUILD_V1=false
      BUILD_V2=true
      shift
      ;;
    --help)
      echo "Usage: $0 [--brand BRAND_NAME] [--v1-only | --v2-only]"
      echo "  --brand BRAND_NAME   Use custom branding (default: delving)"
      echo "  --v1-only           Build only V1 API documentation"
      echo "  --v2-only           Build only V2 API documentation"
      echo ""
      echo "Brands available in $TEMPLATES_DIR/brands/"
      exit 0
      ;;
    *)
      echo "Unknown option: $1"
      exit 1
      ;;
  esac
done

# Create necessary directories
mkdir -p "$PDF_DIR"
mkdir -p "$TEMPLATES_DIR/brands/delving"

# Function to check dependencies
check_dependencies() {
  local missing=()
  
  if ! command -v pandoc &> /dev/null; then
    missing+=("pandoc")
  fi
  
  if ! command -v xelatex &> /dev/null && ! command -v pdflatex &> /dev/null; then
    missing+=("texlive (xelatex or pdflatex)")
  fi
  
  if [ ${#missing[@]} -ne 0 ]; then
    echo "Error: Missing required dependencies: ${missing[*]}"
    echo "Please install them first."
    exit 1
  fi
}

# Function to create default Delving template if it doesn't exist
create_default_template() {
  local template_file="$TEMPLATES_DIR/api-template-simple.tex"
  
  if [ ! -f "$template_file" ]; then
    echo "Creating default API documentation template..."
    cat > "$template_file" <<'EOF'
% Hub3 API Documentation Template
\documentclass[11pt,a4paper]{article}

% Essential packages
\usepackage[utf8]{inputenc}
\usepackage[english]{babel}
\usepackage{graphicx}
\usepackage{fancyhdr}
\usepackage[margin=2.5cm, headheight=2cm, footskip=1.5cm]{geometry}
\usepackage{xcolor}
\usepackage{hyperref}
\usepackage{booktabs}
\usepackage{array}
\usepackage{enumitem}
\usepackage{longtable}
\usepackage{titlesec}
\usepackage{setspace}
\usepackage{calc}
\usepackage{listings}
\usepackage{tcolorbox}
\usepackage{datetime}

% Brand colors (overridden by brand-specific includes)
\definecolor{brandprimary}{RGB}{41, 118, 176}
\definecolor{brandsecondary}{RGB}{25, 85, 140}
\definecolor{lightgray}{RGB}{240, 240, 240}
\definecolor{codebg}{RGB}{248, 248, 248}

% Include brand-specific settings if they exist
\IfFileExists{$BRANDCONFIG$}{%
  \input{$BRANDCONFIG$}
}{}

% Set up hyperref
\hypersetup{
    colorlinks=true,
    linkcolor=brandprimary,
    urlcolor=brandprimary,
    citecolor=brandprimary
}

% Define tightlist for pandoc compatibility
\providecommand{\tightlist}{%
  \setlength{\itemsep}{0pt}\setlength{\parskip}{0pt}}

% Code listing style
\lstset{
  backgroundcolor=\color{codebg},
  basicstyle=\ttfamily\small,
  breaklines=true,
  frame=single,
  frameround=tttt,
  rulecolor=\color{lightgray},
  keywordstyle=\color{brandsecondary}\bfseries,
  stringstyle=\color{brandprimary},
  commentstyle=\color{gray}\itshape,
  showstringspaces=false,
  tabsize=2
}

% Better section formatting
\titleformat{\section}
  {\normalfont\Large\bfseries\color{brandsecondary}}
  {\thesection}{0.5em}{}
  [\titlerule]
  
\titleformat{\subsection}
  {\normalfont\large\bfseries\color{brandsecondary}}
  {\thesubsection}{0.5em}{}

\titleformat{\subsubsection}
  {\normalfont\normalsize\bfseries\color{brandsecondary}}
  {\thesubsubsection}{0.5em}{}

\titlespacing*{\section}{0pt}{1.5\baselineskip}{1\baselineskip}
\titlespacing*{\subsection}{0pt}{1.2\baselineskip}{0.8\baselineskip}
\titlespacing*{\subsubsection}{0pt}{1\baselineskip}{0.5\baselineskip}

% Header and footer setup
\pagestyle{fancy}
\fancyhf{}

% Logo in header - uses brand logo
\fancyhead[L]{\IfFileExists{$BRANDLOGO$}{\includegraphics[height=1.5cm]{$BRANDLOGO$}}{}}
\fancyhead[R]{API Documentation}

% Footer with metadata
\fancyfoot[L]{%
    \scriptsize
    $if(brandname)$$brandname$$else$Hub3 API$endif$\\
    Generated: \today\\
    Git SHA: $gitsha$
}
\fancyfoot[C]{\thepage}
\fancyfoot[R]{%
    \scriptsize
    Version: $apiversion$\\
    Build: $timestamp$
}

% Line under the header
\renewcommand{\headrulewidth}{0.4pt}
\renewcommand{\footrulewidth}{0.4pt}

% Custom title page
\newcommand{\customtitlepage}{
  \thispagestyle{empty}
  \begin{center}
    \vspace*{1cm}
    
    % Logo
    \IfFileExists{$BRANDLOGO$}{%
      \includegraphics[height=3cm]{$BRANDLOGO$}\\[1cm]
    }{}
    
    % Title
    {\fontsize{24}{28}\selectfont\textbf{\textcolor{brandsecondary}{$title$}}}
    
    \vspace{1cm}
    
    % Subtitle
    $if(subtitle)$
    {\Large\textcolor{brandprimary}{$subtitle$}}
    \vspace{0.5cm}
    $endif$
    
    % Version info box
    \begin{tcolorbox}[
      colback=lightgray,
      colframe=brandprimary,
      width=0.7\textwidth,
      arc=2mm,
      boxrule=0.5pt
    ]
      \centering
      \textbf{Version Information}\\[0.3cm]
      API Version: $apiversion$\\
      Build Date: \today\\
      Build Time: $timestamp$\\
      Git SHA: \texttt{$gitsha$}
    \end{tcolorbox}
    
    \vfill
    
    % Organization info
    $if(brandname)$
    {\large $brandname$}\\
    $endif$
    $if(brandwebsite)$
    \url{$brandwebsite$}
    $endif$
    
  \end{center}
  \newpage
}

% Begin document
\begin{document}

% Custom title page
\customtitlepage

% Table of contents
\tableofcontents
\newpage

% Main content
$body$

\end{document}
EOF
    echo "Template created."
  fi
}

# Function to create default Delving brand config
create_delving_brand() {
  local brand_dir="$TEMPLATES_DIR/brands/delving"
  local brand_config="$brand_dir/config.tex"
  
  mkdir -p "$brand_dir"
  
  if [ ! -f "$brand_config" ]; then
    echo "Creating Delving brand configuration..."
    cat > "$brand_config" <<'EOF'
% Delving brand colors
\definecolor{brandprimary}{RGB}{41, 118, 176}
\definecolor{brandsecondary}{RGB}{25, 85, 140}
\definecolor{lightgray}{RGB}{240, 240, 240}
\definecolor{codebg}{RGB}{248, 248, 248}

% Delving specific settings
\def\brandname{Delving B.V.}
\def\brandwebsite{https://www.delving.eu}
EOF
  fi
  
  # Create a placeholder for logo if it doesn't exist
  if [ ! -f "$brand_dir/logo.png" ] && [ ! -f "$brand_dir/logo.eps" ]; then
    echo "Note: Please add a logo file to $brand_dir/logo.png or logo.eps"
  fi
}

# Function to build a single PDF
build_pdf() {
  local input_file="$1"
  local api_version="$2"
  local output_base=$(basename "$input_file" .md | sed 's/_/-/g')
  local output_file="$PDF_DIR/${output_base}-${BRAND}-${TIMESTAMP}-${GIT_SHA}.pdf"
  
  echo "Building $output_base PDF..."
  
  # Determine brand logo path
  local brand_logo=""
  if [ -f "$TEMPLATES_DIR/brands/$BRAND/logo.eps" ]; then
    brand_logo="$TEMPLATES_DIR/brands/$BRAND/logo.eps"
  elif [ -f "$TEMPLATES_DIR/brands/$BRAND/logo.png" ]; then
    brand_logo="$TEMPLATES_DIR/brands/$BRAND/logo.png"
  fi
  
  # Build with pandoc
  if [ -n "$brand_logo" ]; then
    pandoc "$input_file" \
      --template="$TEMPLATES_DIR/api-template-simple.tex" \
      --pdf-engine=xelatex \
      -V BRANDCONFIG="$TEMPLATES_DIR/brands/$BRAND/config.tex" \
      -V BRANDLOGO="$brand_logo" \
      -V apiversion="$api_version" \
      -V timestamp="$TIMESTAMP" \
      -V gitsha="$GIT_SHA" \
      -V title="Hub3 API $api_version Documentation" \
      -V subtitle="RESTful API Reference Guide" \
      --toc \
      --toc-depth=3 \
      --highlight-style=tango \
      --listings \
      -o "$output_file"
  else
    pandoc "$input_file" \
      --template="$TEMPLATES_DIR/api-template-simple.tex" \
      --pdf-engine=xelatex \
      -V BRANDCONFIG="$TEMPLATES_DIR/brands/$BRAND/config.tex" \
      -V apiversion="$api_version" \
      -V timestamp="$TIMESTAMP" \
      -V gitsha="$GIT_SHA" \
      -V title="Hub3 API $api_version Documentation" \
      -V subtitle="RESTful API Reference Guide" \
      --toc \
      --toc-depth=3 \
      --highlight-style=tango \
      --listings \
      -o "$output_file"
  fi
  
  if [ $? -eq 0 ]; then
    echo "✓ Success! PDF created at $output_file"
    
    # Create a symlink to latest version
    local latest_link="$PDF_DIR/${output_base}-${BRAND}-latest.pdf"
    ln -sf "$(basename "$output_file")" "$latest_link"
    echo "✓ Latest link created at $latest_link"
  else
    echo "✗ Error: PDF generation failed for $input_file"
    return 1
  fi
}

# Main execution
echo "Hub3 API Documentation PDF Builder"
echo "=================================="
echo "Brand: $BRAND"
echo "Timestamp: $TIMESTAMP"
echo "Git SHA: $GIT_SHA"
echo ""

# Check dependencies
check_dependencies

# Create templates and brand configs
create_default_template
create_delving_brand

# Check if brand exists
if [ "$BRAND" != "delving" ] && [ ! -d "$TEMPLATES_DIR/brands/$BRAND" ]; then
  echo "Warning: Brand '$BRAND' not found. Creating brand directory..."
  mkdir -p "$TEMPLATES_DIR/brands/$BRAND"
  echo "Please add the following files to $TEMPLATES_DIR/brands/$BRAND/:"
  echo "  - config.tex (brand colors and settings)"
  echo "  - logo.png or logo.eps (brand logo)"
  echo ""
  echo "Example config.tex:"
  echo "  \definecolor{brandprimary}{RGB}{0, 0, 255}"
  echo "  \definecolor{brandsecondary}{RGB}{0, 0, 128}"
  echo "  \def\brandname{Your Company Name}"
  echo "  \def\brandwebsite{https://www.example.com}"
  echo ""
fi

# Build PDFs
if [ "$BUILD_V1" = true ] && [ -f "V1_API_DOCUMENTATION.md" ]; then
  build_pdf "V1_API_DOCUMENTATION.md" "v1"
fi

if [ "$BUILD_V2" = true ] && [ -f "V2_API_DOCUMENTATION.md" ]; then
  build_pdf "V2_API_DOCUMENTATION.md" "v2"
fi

echo ""
echo "Build complete!"
echo "PDFs are available in the $PDF_DIR directory."