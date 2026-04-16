package download

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestFetchLatestTag(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		statusCode int
		wantTag    string
		wantErr    string
	}{
		{
			name:       "returns tag from valid response",
			body:       `{"tag_name":"v1.2.3","assets":[]}`,
			statusCode: http.StatusOK,
			wantTag:    "v1.2.3",
		},
		{
			name:       "returns error on non-200",
			body:       `{"message":"Not Found"}`,
			statusCode: http.StatusNotFound,
			wantErr:    "GitHub API error",
		},
		{
			name:       "returns error on invalid JSON",
			body:       `not-json`,
			statusCode: http.StatusOK,
			wantErr:    "decode release JSON",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.body)) //nolint:errcheck
			}))
			defer server.Close()

			// Swap out the URL by using a client that routes to the test server.
			// We monkey-patch FetchLatestRelease by calling it directly with a custom URL approach.
			// Instead, test via a custom RoundTripper that intercepts github.com requests.
			transport := &fixedResponseTransport{server: server}
			client := &http.Client{Transport: transport}

			tag, err := FetchLatestTag(context.Background(), client, "owner", "repo")
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("error = %v, want substring %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("FetchLatestTag() error = %v", err)
			}
			if tag != tt.wantTag {
				t.Fatalf("tag = %q, want %q", tag, tt.wantTag)
			}
		})
	}
}

// fixedResponseTransport redirects all requests to a test httptest.Server.
type fixedResponseTransport struct {
	server *httptest.Server
}

func (t *fixedResponseTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	cloned := req.Clone(req.Context())
	cloned.URL.Scheme = "http"
	cloned.URL.Host = t.server.Listener.Addr().String()
	return http.DefaultTransport.RoundTrip(cloned)
}
