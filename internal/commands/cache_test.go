package commands

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCollectURLs_SingleURL(t *testing.T) {
	urls, err := collectURLs("https://example.com/img.jpg", "", "", false, "", "")
	if err != nil {
		t.Fatal(err)
	}
	if len(urls) != 1 {
		t.Fatalf("got %d urls, want 1", len(urls))
	}
	if urls[0] != "https://example.com/img.jpg" {
		t.Errorf("got %q, want %q", urls[0], "https://example.com/img.jpg")
	}
}

func TestCollectURLs_CommaSeparated(t *testing.T) {
	urls, err := collectURLs("", "https://a.com/1.jpg,https://b.com/2.jpg", "", false, "", "")
	if err != nil {
		t.Fatal(err)
	}
	if len(urls) != 2 {
		t.Fatalf("got %d urls, want 2", len(urls))
	}
}

func TestCollectURLs_FromFile(t *testing.T) {
	tmpDir := t.TempDir()
	f := filepath.Join(tmpDir, "urls.txt")
	os.WriteFile(f, []byte("https://a.com/1.jpg\nhttps://b.com/2.jpg\nhttps://c.com/3.jpg\n"), 0644)

	urls, err := collectURLs("", "", f, false, "", "")
	if err != nil {
		t.Fatal(err)
	}
	if len(urls) != 3 {
		t.Fatalf("got %d urls, want 3", len(urls))
	}
}

func TestCollectURLs_Deduplication(t *testing.T) {
	urls, err := collectURLs("https://a.com/1.jpg", "https://a.com/1.jpg,https://b.com/2.jpg", "", false, "", "")
	if err != nil {
		t.Fatal(err)
	}
	if len(urls) != 2 {
		t.Fatalf("got %d urls, want 2 (deduped)", len(urls))
	}
}

func TestCollectURLs_PathWithSubdomain(t *testing.T) {
	urls, err := collectURLs("", "", "", false, "/products/test.jpg", "mysub")
	if err != nil {
		t.Fatal(err)
	}
	if len(urls) != 1 {
		t.Fatalf("got %d urls, want 1", len(urls))
	}
	want := "https://mysub.gumlet.io/products/test.jpg"
	if urls[0] != want {
		t.Errorf("got %q, want %q", urls[0], want)
	}
}

func TestCollectURLs_PathWithoutSubdomain(t *testing.T) {
	_, err := collectURLs("", "", "", false, "/test.jpg", "")
	if err == nil {
		t.Fatal("expected error for --path without --subdomain")
	}
}

func TestCollectURLs_Empty(t *testing.T) {
	urls, err := collectURLs("", "", "", false, "", "")
	if err != nil {
		t.Fatal(err)
	}
	if len(urls) != 0 {
		t.Fatalf("got %d urls, want 0", len(urls))
	}
}

func TestExtractSubdomain(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"https://millefarmacie.gumlet.io/img.jpg", "millefarmacie"},
		{"https://other.gumlet.io/path/img.jpg", "other"},
		{"https://example.com/img.jpg", ""},
		{"http://sub.gumlet.io/test", "sub"},
	}
	for _, tt := range tests {
		got := extractSubdomain(tt.input)
		if got != tt.want {
			t.Errorf("extractSubdomain(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
