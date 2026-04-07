package download

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type GitHubRelease struct {
	TagName string        `json:"tag_name"`
	Assets  []GitHubAsset `json:"assets"`
}

type GitHubAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
	Size               int64  `json:"size"`
	ContentType        string `json:"content_type"`
}

func FetchLatestRelease(ctx context.Context, client *http.Client, owner, repo string) (*GitHubRelease, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", owner, repo)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("GitHub API unreachable: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return nil, fmt.Errorf("GitHub API error: %s — %s", resp.Status, strings.TrimSpace(string(body)))
	}

	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("decode release JSON: %w", err)
	}
	return &release, nil
}

func SelectAsset(release *GitHubRelease, goos, goarch string) (archiveAsset *GitHubAsset, checksumAsset *GitHubAsset, err error) {
	pattern := AssetPattern(goos, goarch)
	ext := ArchiveExt(goos)

	for i := range release.Assets {
		asset := &release.Assets[i]
		name := asset.Name
		if strings.Contains(name, pattern) && strings.HasSuffix(name, ext) {
			archiveAsset = asset
		}
		if name == "checksums.txt" || strings.HasSuffix(name, ".sha256") {
			checksumAsset = asset
		}
	}

	if archiveAsset == nil {
		return nil, nil, fmt.Errorf("no engram release asset found for %s/%s", goos, goarch)
	}
	return archiveAsset, checksumAsset, nil
}
