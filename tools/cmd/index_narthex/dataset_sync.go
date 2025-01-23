package main

import (
	"archive/zip"
	"encoding/xml"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

const (
	maxRetries    = 3
	retryInterval = 5 * time.Second
)

type SipZips struct {
	XMLName   xml.Name  `xml:"sip-zips"`
	Available Available `xml:"available"`
}

type Available struct {
	SipZip []SipZip `xml:"sip-zip"`
}

type SipZip struct {
	Dataset string `xml:"dataset"`
	File    string `xml:"file"`
}

type FailedDownload struct {
	Dataset string
	File    string
	Error   error
}

type HarvestConfig struct {
	URL      string
	Username string
	Password string
	Output   string
	OrgID    string
}

func fetchXML(config HarvestConfig) (*SipZips, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", config.URL, nil)
	if err != nil {
		return nil, err
	}

	if config.Username != "" || config.Password != "" {
		req.SetBasicAuth(config.Username, config.Password)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP error: %s", resp.Status)
	}

	var sipZips SipZips
	if err := xml.NewDecoder(resp.Body).Decode(&sipZips); err != nil {
		return nil, err
	}

	return &sipZips, nil
}

func downloadFile(url, filepath, username, password string) error {
	var lastErr error
	for attempt := 1; attempt <= maxRetries; attempt++ {
		if err := downloadFileOnce(url, filepath, username, password); err != nil {
			lastErr = err
			slog.Warn("download attempt failed",
				"attempt", attempt,
				"url", url,
				"error", err)
			if attempt < maxRetries {
				time.Sleep(retryInterval)
				continue
			}
			return fmt.Errorf("all download attempts failed: %w", lastErr)
		}
		return nil
	}
	return lastErr
}

func downloadFileOnce(url, filepath, username, password string) error {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	if username != "" || password != "" {
		req.SetBasicAuth(username, password)
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP error: %s", resp.Status)
	}

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

func processFile(config HarvestConfig, dataset, filename string) error {
	datasetDir := filepath.Join(config.Output, dataset)
	if err := os.MkdirAll(datasetDir, 0o755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	zipPath := filepath.Join(datasetDir, filename)
	defer os.Remove(zipPath) // Clean up zip file after processing

	downloadURL := fmt.Sprintf("%s/%s", config.URL, dataset)
	slog.Info("downloading file", "url", downloadURL, "output", zipPath)

	if err := downloadFile(downloadURL, zipPath, config.Username, config.Password); err != nil {
		return fmt.Errorf("failed to download file: %w", err)
	}

	if err := unzipFile(zipPath, datasetDir); err != nil {
		return fmt.Errorf("failed to unzip file: %w", err)
	}

	err := filepath.Walk(datasetDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		basename := filepath.Base(path)
		if strings.HasPrefix(basename, "mapping_") || strings.HasSuffix(basename, "_record-definition.xml") {
			if err := processTextFile(path, config.OrgID); err != nil {
				return fmt.Errorf("failed to process %s: %w", basename, err)
			}
		}
		return nil
	})

	return err
}

func ProcessAllFiles(config HarvestConfig) error {
	sipZips, err := fetchXML(config)
	if err != nil {
		return fmt.Errorf("failed to fetch XML: %w", err)
	}

	var failedDownloads []FailedDownload

	// First pass - try to process all files
	for _, sipZip := range sipZips.Available.SipZip {
		if err := processFile(config, sipZip.Dataset, sipZip.File); err != nil {
			slog.Error("failed to process file",
				"dataset", sipZip.Dataset,
				"file", sipZip.File,
				"error", err)
			failedDownloads = append(failedDownloads, FailedDownload{
				Dataset: sipZip.Dataset,
				File:    sipZip.File,
				Error:   err,
			})
		}
	}

	// Retry failed downloads
	if len(failedDownloads) > 0 {
		slog.Info("retrying failed downloads", "count", len(failedDownloads))
		var stillFailed []FailedDownload

		for _, fd := range failedDownloads {
			if err := processFile(config, fd.Dataset, fd.File); err != nil {
				slog.Error("retry failed",
					"dataset", fd.Dataset,
					"file", fd.File,
					"error", err)
				stillFailed = append(stillFailed, fd)
			}
		}

		if len(stillFailed) > 0 {
			for _, fd := range stillFailed {
				slog.Error("permanently failed download",
					"dataset", fd.Dataset,
					"file", fd.File,
					"error", fd.Error)
			}
			return fmt.Errorf("%d downloads failed permanently", len(stillFailed))
		}
	}

	return nil
}

func unzipFile(zipPath, destDir string) error {
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer reader.Close()

	for _, file := range reader.File {
		if err := extractZipFile(file, destDir); err != nil {
			return err
		}
	}
	return nil
}

func extractZipFile(file *zip.File, destDir string) error {
	rc, err := file.Open()
	if err != nil {
		return err
	}
	defer rc.Close()

	path := filepath.Join(destDir, file.Name)
	if file.FileInfo().IsDir() {
		os.MkdirAll(path, file.Mode())
		return nil
	}

	os.MkdirAll(filepath.Dir(path), 0o755)
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, rc)
	return err
}

func replaceOrgID(data, orgID string) string {
	oldStr := `<entry>
     <string>orgId</string>
     <string></string>
   </entry>`
	newStr := fmt.Sprintf(`<entry>
     <string>orgId</string>
     <string>%s</string>
   </entry>`, orgID)

	return strings.ReplaceAll(string(data), oldStr, newStr)
}

func processTextFile(filepath, orgID string) error {
	content, err := os.ReadFile(filepath)
	if err != nil {
		return err
	}
	patterns := []struct {
		pattern     *regexp.Regexp
		replacement string
	}{
		{regexp.MustCompile(`edm:RDF`), "rdf:RDF"},
		{regexp.MustCompile(`/RDF/`), "/rdf:RDF/"},
		{regexp.MustCompile(`root tag="RDF"`), `root tag="rdf:RDF"`},
		{
			regexp.MustCompile(`(?s)<entry>\s*<string>orgId</string>\s*<string>\s*</string>\s*</entry>`),
			fmt.Sprintf("<entry>\n      <string>orgId</string>\n      <string>%s</string>\n    </entry>", orgID),
		},
	}

	modified := string(content)
	for _, p := range patterns {
		modified = p.pattern.ReplaceAllString(modified, p.replacement)
	}

	modified = replaceOrgID(modified, orgID)

	return os.WriteFile(filepath, []byte(modified), 0o644)
}
