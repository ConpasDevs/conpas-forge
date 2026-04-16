package checker

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/conpasDEVS/conpas-forge/internal/config"
)

func makeRelease(tag string) string {
	return `{"tag_name":"` + tag + `","assets":[]}`
}

func TestCheckVersions_Statuses(t *testing.T) {
	tests := []struct {
		name            string
		serverResponse  string
		serverStatus    int
		engramInstalled bool
		engramVersion   string
		gaInstalled     bool
		gaVersion       string
		wantEngram      string
		wantForge       string
	}{
		{
			name:            "both up-to-date",
			serverResponse:  makeRelease("v1.0.0"),
			serverStatus:    http.StatusOK,
			engramInstalled: true,
			engramVersion:   "v1.0.0",
			gaInstalled:     true,
			gaVersion:       "v1.0.0",
			wantEngram:      StatusUpToDate,
			// version.Version == "dev" in test builds → compareVersions("dev","v1.0.0") → unknown
			wantForge: StatusUnknown,
		},
		{
			name:            "engram outdated",
			serverResponse:  makeRelease("v2.0.0"),
			serverStatus:    http.StatusOK,
			engramInstalled: true,
			engramVersion:   "v1.0.0",
			gaInstalled:     true,
			gaVersion:       "v2.0.0",
			wantEngram:      StatusOutdated,
			// version.Version == "dev" in test builds → unknown
			wantForge: StatusUnknown,
		},
		{
			name:            "engram not installed",
			serverResponse:  makeRelease("v1.0.0"),
			serverStatus:    http.StatusOK,
			engramInstalled: false,
			engramVersion:   "",
			gaInstalled:     true,
			gaVersion:       "v1.0.0",
			wantEngram:      StatusNotInstalled,
			// version.Version == "dev" in test builds → unknown
			wantForge: StatusUnknown,
		},
		{
			name:            "api offline returns unknown",
			serverResponse:  "",
			serverStatus:    http.StatusInternalServerError,
			engramInstalled: true,
			engramVersion:   "v1.0.0",
			gaInstalled:     true,
			gaVersion:       "v1.0.0",
			wantEngram:      StatusUnknown,
			wantForge:       StatusUnknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.serverStatus)
				if tt.serverResponse != "" {
					w.Write([]byte(tt.serverResponse)) //nolint:errcheck
				}
			}))
			defer server.Close()

			// Use custom transport to redirect all GitHub API calls to test server
			transport := &redirectTransport{target: server.URL}
			client := &http.Client{Transport: transport}

			cfg := &config.Config{
				Modules: config.ModulesConfig{
					Engram: config.ModuleStatus{
						Installed: tt.engramInstalled,
						Version:   tt.engramVersion,
					},
					GentleAI: config.ModuleStatus{
						Installed: tt.gaInstalled,
						Version:   tt.gaVersion,
					},
				},
			}

			results, err := CheckVersions(context.Background(), client, cfg)
			if err != nil {
				t.Fatalf("CheckVersions() error = %v", err)
			}
			if len(results) != 2 {
				t.Fatalf("len(results) = %d, want 2", len(results))
			}

			engramResult := results[0]
			forgeResult := results[1]

			if engramResult.Status != tt.wantEngram {
				t.Errorf("Engram status = %q, want %q", engramResult.Status, tt.wantEngram)
			}
			if forgeResult.Status != tt.wantForge {
				t.Errorf("conpas-forge status = %q, want %q", forgeResult.Status, tt.wantForge)
			}
		})
	}
}

func TestCheckVersions_ModuleNames(t *testing.T) {
	cfg := &config.Config{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(makeRelease("v1.0.0"))) //nolint:errcheck
	}))
	defer server.Close()

	transport := &redirectTransport{target: server.URL}
	client := &http.Client{Transport: transport}

	results, err := CheckVersions(context.Background(), client, cfg)
	if err != nil {
		t.Fatalf("CheckVersions() error = %v", err)
	}
	if results[0].Module != "Engram" {
		t.Errorf("module[0] = %q, want %q", results[0].Module, "Engram")
	}
	if results[1].Module != "conpas-forge" {
		t.Errorf("module[1] = %q, want %q", results[1].Module, "conpas-forge")
	}
}

// TestCheckVersions_ConpasForgeNeverNotInstalled verifies that conpas-forge is never
// reported as "not-installed", even when the config marks GentleAI as not installed.
func TestCheckVersions_ConpasForgeNeverNotInstalled(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(makeRelease("v1.0.0"))) //nolint:errcheck
	}))
	defer server.Close()

	transport := &redirectTransport{target: server.URL}
	client := &http.Client{Transport: transport}

	// GentleAI not installed, no version in config
	cfg := &config.Config{
		Modules: config.ModulesConfig{
			GentleAI: config.ModuleStatus{Installed: false, Version: ""},
		},
	}

	results, err := CheckVersions(context.Background(), client, cfg)
	if err != nil {
		t.Fatalf("CheckVersions() error = %v", err)
	}
	forgeResult := results[1]
	if forgeResult.Status == StatusNotInstalled {
		t.Errorf("conpas-forge status = %q; conpas-forge is always installed (it IS the running binary)", forgeResult.Status)
	}
}

// redirectTransport rewrites all HTTP requests to a fixed target host, preserving the path.
type redirectTransport struct {
	target string
}

func (t *redirectTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	cloned := req.Clone(req.Context())
	cloned.URL.Scheme = "http"
	cloned.URL.Host = req.Host
	// Replace host with test server
	serverURL := t.target
	if len(serverURL) > 7 && serverURL[:7] == "http://" {
		cloned.URL.Host = serverURL[7:]
	}
	return http.DefaultTransport.RoundTrip(cloned)
}
