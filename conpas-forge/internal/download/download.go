package download

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

func DownloadToTemp(ctx context.Context, client *http.Client, url string, onProgress func(bytesRead, total int64)) (tmpPath string, err error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("create download request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("download request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download HTTP error: %s", resp.Status)
	}

	tmp, err := os.CreateTemp("", "conpas-forge-dl-*")
	if err != nil {
		return "", fmt.Errorf("create temp file: %w", err)
	}

	pr := &ProgressReader{
		Reader:     resp.Body,
		Total:      resp.ContentLength,
		OnProgress: onProgress,
	}

	if _, err := io.Copy(tmp, pr); err != nil {
		tmp.Close()
		os.Remove(tmp.Name())
		return "", fmt.Errorf("download stream failed: %w", err)
	}

	if err := tmp.Close(); err != nil {
		os.Remove(tmp.Name())
		return "", fmt.Errorf("close temp file: %w", err)
	}

	return tmp.Name(), nil
}

func VerifyChecksum(filePath, expectedHex string) error {
	f, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("open file for checksum: %w", err)
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return fmt.Errorf("compute checksum: %w", err)
	}

	computed := hex.EncodeToString(h.Sum(nil))
	if !strings.EqualFold(computed, strings.TrimSpace(expectedHex)) {
		return fmt.Errorf("checksum mismatch: expected %s, got %s", expectedHex, computed)
	}
	return nil
}

func FetchChecksumHex(ctx context.Context, client *http.Client, asset *GitHubAsset, archiveAssetName string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", asset.BrowserDownloadURL, nil)
	if err != nil {
		return "", fmt.Errorf("create checksum request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("download checksum: %w", err)
	}
	defer resp.Body.Close()

	var lines []string
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("scan checksum file: %w", err)
	}

	// Multi-line checksums.txt: find the line containing the archive name
	for _, line := range lines {
		if strings.Contains(line, archiveAssetName) {
			parts := strings.Fields(line)
			if len(parts) >= 1 {
				return parts[0], nil
			}
		}
	}

	// Bare-hex fallback: single-line file containing only the hex digest
	if len(lines) == 1 {
		hex := strings.TrimSpace(lines[0])
		if len(hex) == 64 && !strings.ContainsAny(hex, " \t/") {
			return hex, nil
		}
	}

	return "", fmt.Errorf("checksum for %q not found in asset", archiveAssetName)
}
