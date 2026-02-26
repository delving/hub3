// Command gendocs generates the Semantic API reference documentation
// from code, ensuring the documentation never drifts from the implementation.
//
// Usage:
//
//	go run ./tools/cmd/gendocs                           # write to docs/semantic-api-reference.md
//	go run ./tools/cmd/gendocs -o /path/to/output.md     # write to custom path
//	go run ./tools/cmd/gendocs -base /api/semantic/v1    # set base URL
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	semantic "github.com/delving/hub3/ikuzo/service/x/semantic"
)

func main() {
	var (
		output  string
		baseURL string
	)

	flag.StringVar(&output, "o", "", "Output file path (default: docs/semantic-api-reference.md)")
	flag.StringVar(&baseURL, "base", "/api/semantic/v1", "API base URL for examples")
	flag.Parse()

	if output == "" {
		// Find project root by looking for go.mod
		dir, err := os.Getwd()
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}

		for {
			if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
				break
			}

			parent := filepath.Dir(dir)
			if parent == dir {
				fmt.Fprintln(os.Stderr, "error: could not find go.mod in parent directories")
				os.Exit(1)
			}

			dir = parent
		}

		output = filepath.Join(dir, "docs", "semantic-api-reference.md")
	}

	markdown := semantic.GenerateAPIReference(baseURL)

	if err := os.MkdirAll(filepath.Dir(output), 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "error creating directory: %v\n", err)
		os.Exit(1)
	}

	if err := os.WriteFile(output, []byte(markdown), 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "error writing file: %v\n", err)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "Generated %s\n", output)
}
