package essync

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// calculateChecksum generates a SHA-256 checksum of the document
// after normalizing field order.
func calculateChecksum(data map[string]interface{}) string {
	// Remove existing checksum if present
	delete(data, "_checksum")

	// Sort keys for consistent checksum
	keys := make([]string, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Create ordered map
	ordered := make(map[string]interface{})
	for _, k := range keys {
		ordered[k] = data[k]
	}

	bytes, _ := json.Marshal(ordered)
	return Checksum(bytes)
}

func Checksum(bytes []byte) string {
	hash := sha256.Sum256(bytes)
	return hex.EncodeToString(hash[:])
}

func FindLatestPath(path string, searchStr string) (string, error) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return "", err
	}

	// If it's a file and contains the search string, return it
	if !fileInfo.IsDir() {
		if strings.Contains(path, searchStr) {
			return path, nil
		}
		return "", nil
	}

	// For directory, find all files
	files, err := os.ReadDir(path)
	if err != nil {
		return "", err
	}

	var latestTime time.Time
	var latestFile string

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		if !strings.Contains(file.Name(), searchStr) {
			continue
		}

		name := file.Name()
		parts := strings.Split(name, "__")
		if len(parts) < 2 {
			continue
		}

		// Try parsing the date part - support multiple formats
		dateStr := parts[0]
		var date time.Time
		var err error
		
		// Try different timestamp formats
		formats := []string{
			"2006-01-02T15:04:05",  // Original format with colons
			"2006-01-02T150405",    // Format without colons (like yours)
			"2006-01-02T15:04",     // Format without seconds
			"2006-01-02",           // Date only
		}
		
		for _, format := range formats {
			date, err = time.Parse(format, dateStr)
			if err == nil {
				break
			}
		}
		
		if err != nil {
			continue
		}

		if date.After(latestTime) {
			latestTime = date
			latestFile = name
		}
	}

	if latestFile == "" {
		return "", nil
	}

	return filepath.Join(path, latestFile), nil
}
