# Source Analyzer

The Source Analyzer is a tool for analyzing XML documents to understand their structure, content patterns, and generate statistics about the data.

## Installation

```bash
go build -o source_analyzer cmd/analyzer/main.go
```

## Usage

The basic command structure is:

```bash
source_analyzer analyze [flags]
```

### Flags

- `-i, --input`: Input XML file to analyze (required)
- `-o, --output`: Output directory for analysis results (required)
- `-c, --compress-output`: Compress output files using zstd

### Example Usage

#### Basic Analysis
```bash
./source_analyzer analyze -i input.xml -o /tmp/analysis_output
```

#### Using Compressed Input
```bash
# First compress your XML file
zstd input.xml -o input.xml.zst

# Analyze compressed file
./source_analyzer analyze -i input.xml.zst -o /tmp/analysis_output
```

#### Generate Compressed Output
```bash
./source_analyzer analyze -i input.xml -o /tmp/analysis_output -c
```

## Output Structure

The analyzer creates a directory structure containing various analysis files:

```
/tmp/analysis_output/
└── tree/
    ├── root/
    │   ├── record/
    │   │   ├── title/
    │   │   │   ├── values.txt          # Raw values
    │   │   │   ├── sorted.txt          # Sorted values
    │   │   │   ├── histogram.txt       # Value frequency data
    │   │   │   ├── histogram-100.json  # Top 100 values histogram
    │   │   │   └── sample-100.json     # Random sample of 100 values
    │   │   ├── description/
    │   │   ├── tags/
    │   │   └── metadata/
    │   └── status.json
    └── status.json
```

### Output Files

- `values.txt`: Contains all raw values for a node
- `sorted.txt`: Values sorted alphabetically
- `histogram.txt`: Frequency count of values
- `histogram-{size}.json`: JSON representation of value frequencies
- `sample-{size}.json`: Random sample of values
- `status.json`: Processing status and statistics

## Testing

Here's a simple test XML file you can use to try out the analyzer:

```xml
<?xml version="1.0" encoding="UTF-8"?>
<root>
    <record id="1">
        <title>First Record</title>
        <description>This is a test description</description>
        <tags>
            <tag>test</tag>
            <tag>example</tag>
        </tags>
        <metadata date="2024-02-19">
            <author>John Doe</author>
            <category>Test</category>
        </metadata>
    </record>
    <record id="2">
        <title>Second Record</title>
        <description>Another test description</description>
        <tags>
            <tag>sample</tag>
            <tag>test</tag>
        </tags>
        <metadata date="2024-02-19">
            <author>Jane Smith</author>
            <category>Example</category>
        </metadata>
    </record>
</root>
```

Save this as `test.xml` and run:

```bash
./source_analyzer analyze -i test.xml -o /tmp/analysis_output
```

### Verifying Results

Check the analysis output:

```bash
# View histograms
cat /tmp/analysis_output/tree/root/record/title/histogram-100.json

# View samples
cat /tmp/analysis_output/tree/root/record/title/sample-100.json

# View status
cat /tmp/analysis_output/tree/root/status.json
```

## Features

- XML parsing with namespace support
- Compression support (zstd)
- Value frequency analysis
- Random sampling
- Histogram generation
- Progress reporting
- Concurrent processing
- Memory-efficient handling of large files

## Notes

- The analyzer uses temporary files for sorting large datasets
- All paths are automatically sanitized for filesystem compatibility
- Progress is reported every second during processing
- Memory usage is optimized for large files through chunked processing
