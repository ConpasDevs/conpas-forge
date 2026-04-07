package download

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSelectAssetPrefersArchiveSpecificChecksum(t *testing.T) {
	tests := []struct {
		name             string
		release          *GitHubRelease
		wantArchive      string
		wantChecksumName string
	}{
		{
			name: "prefers archive specific sha over generic checksums file",
			release: &GitHubRelease{Assets: []GitHubAsset{
				{Name: "checksums.txt"},
				{Name: "engram_windows_amd64.zip.sha256"},
				{Name: "engram_windows_amd64.zip"},
			}},
			wantArchive:      "engram_windows_amd64.zip",
			wantChecksumName: "engram_windows_amd64.zip.sha256",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			archive, checksum, err := SelectAsset(tt.release, "windows", "amd64")
			if err != nil {
				t.Fatalf("SelectAsset() error = %v", err)
			}
			if archive.Name != tt.wantArchive {
				t.Fatalf("archive = %q, want %q", archive.Name, tt.wantArchive)
			}
			if checksum == nil || checksum.Name != tt.wantChecksumName {
				t.Fatalf("checksum = %v, want %q", checksum, tt.wantChecksumName)
			}
		})
	}
}

func TestFetchChecksumHex(t *testing.T) {
	validHex := strings.Repeat("a", 64)
	tests := []struct {
		name             string
		statusCode       int
		body             string
		assetName        string
		archiveAssetName string
		want             string
		wantErr          string
	}{
		{
			name:             "rejects non-200 responses",
			statusCode:       http.StatusNotFound,
			body:             "missing",
			assetName:        "engram_windows_amd64.zip.sha256",
			archiveAssetName: "engram_windows_amd64.zip",
			wantErr:          "checksum HTTP error",
		},
		{
			name:             "accepts bare hex only for archive specific checksum files",
			statusCode:       http.StatusOK,
			body:             validHex + "\n",
			assetName:        "engram_windows_amd64.zip.sha256",
			archiveAssetName: "engram_windows_amd64.zip",
			want:             validHex,
		},
		{
			name:             "rejects bare hex from generic checksum assets",
			statusCode:       http.StatusOK,
			body:             validHex + "\n",
			assetName:        "checksums.txt",
			archiveAssetName: "engram_windows_amd64.zip",
			wantErr:          "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				fmt.Fprint(w, tt.body)
			}))
			defer server.Close()

			asset := &GitHubAsset{Name: tt.assetName, BrowserDownloadURL: server.URL}
			got, err := FetchChecksumHex(context.Background(), server.Client(), asset, tt.archiveAssetName)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("error = %v, want substring %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("FetchChecksumHex() error = %v", err)
			}
			if got != tt.want {
				t.Fatalf("checksum = %q, want %q", got, tt.want)
			}
		})
	}
}
