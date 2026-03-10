package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGet_AuthHeader(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != "Bearer test-key" {
			t.Errorf("got auth %q, want %q", auth, "Bearer test-key")
		}
		w.Write([]byte(`{"ok": true}`))
	}))
	defer srv.Close()

	c := New("test-key", srv.URL, false)
	_, err := c.Get(context.Background(), "/test", nil)
	if err != nil {
		t.Fatal(err)
	}
}

func TestGet_AcceptHeader(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		accept := r.Header.Get("Accept")
		if accept != "application/json" {
			t.Errorf("got accept %q, want %q", accept, "application/json")
		}
		w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	c := New("key", srv.URL, false)
	c.Get(context.Background(), "/test", nil)
}

func TestPost_ContentType(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ct := r.Header.Get("Content-Type")
		if ct != "application/json" {
			t.Errorf("got content-type %q, want %q", ct, "application/json")
		}
		w.Write([]byte(`{"created": true}`))
	}))
	defer srv.Close()

	c := New("key", srv.URL, false)
	c.Post(context.Background(), "/test", map[string]string{"name": "test"})
}

func TestGet_404(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		w.Write([]byte(`{"message": "not found"}`))
	}))
	defer srv.Close()

	c := New("key", srv.URL, false)
	_, err := c.Get(context.Background(), "/missing", nil)
	if err == nil {
		t.Fatal("expected error for 404")
	}
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if apiErr.StatusCode != 404 {
		t.Errorf("got status %d, want 404", apiErr.StatusCode)
	}
}

func TestGet_401(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(401)
		w.Write([]byte(`{"error": "unauthorized"}`))
	}))
	defer srv.Close()

	c := New("bad-key", srv.URL, false)
	_, err := c.Get(context.Background(), "/test", nil)
	if err == nil {
		t.Fatal("expected error for 401")
	}
	apiErr := err.(*APIError)
	if apiErr.ExitCode() != 3 {
		t.Errorf("got exit code %d, want 3", apiErr.ExitCode())
	}
}

func TestDeleteNoContent(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("got method %s, want DELETE", r.Method)
		}
		w.WriteHeader(204)
	}))
	defer srv.Close()

	c := New("key", srv.URL, false)
	err := c.Delete(context.Background(), "/resource/123")
	if err != nil {
		t.Fatal(err)
	}
}

func TestBuildURL(t *testing.T) {
	c := New("key", "https://api.gumlet.com", false)

	tests := []struct {
		path string
		want string
	}{
		{"v1/image/source", "https://api.gumlet.com/v1/image/source"},
		{"/v1/image/source", "https://api.gumlet.com/v1/image/source"},
		{"https://other.com/path", "https://other.com/path"},
	}
	for _, tt := range tests {
		got := c.buildURL(tt.path, nil)
		if got != tt.want {
			t.Errorf("buildURL(%q) = %q, want %q", tt.path, got, tt.want)
		}
	}
}

func TestRetryDelay(t *testing.T) {
	// With Retry-After header
	d := retryDelay(0, "5")
	if d != 5e9 {
		t.Errorf("got %v, want 5s", d)
	}

	// Without Retry-After, first attempt should be ~1s
	d = retryDelay(0, "")
	if d < 1e9 || d > 1500000000 {
		t.Errorf("got %v, want ~1s", d)
	}
}

func TestHumanBytes(t *testing.T) {
	tests := []struct {
		input int
		want  string
	}{
		{500, "500B"},
		{1024, "1.0KB"},
		{1536, "1.5KB"},
		{1048576, "1.0MB"},
	}
	for _, tt := range tests {
		got := humanBytes(tt.input)
		if got != tt.want {
			t.Errorf("humanBytes(%d) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
