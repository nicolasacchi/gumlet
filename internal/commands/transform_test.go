package commands

import (
	"testing"
)

func TestBuildTransformURL_WidthOnly(t *testing.T) {
	result, err := buildTransformURL("https://sub.gumlet.io/img.jpg", TransformOptions{Width: 400})
	if err != nil {
		t.Fatal(err)
	}
	want := "https://sub.gumlet.io/img.jpg?w=400"
	if result != want {
		t.Errorf("got %q, want %q", result, want)
	}
}

func TestBuildTransformURL_MultipleParams(t *testing.T) {
	result, err := buildTransformURL("https://sub.gumlet.io/img.jpg", TransformOptions{
		Width:   800,
		Format:  "webp",
		Quality: 80,
	})
	if err != nil {
		t.Fatal(err)
	}
	// URL query params are sorted alphabetically by Go's url.Values.Encode()
	want := "https://sub.gumlet.io/img.jpg?format=webp&q=80&w=800"
	if result != want {
		t.Errorf("got %q, want %q", result, want)
	}
}

func TestBuildTransformURL_AllParams(t *testing.T) {
	result, err := buildTransformURL("https://sub.gumlet.io/img.jpg", TransformOptions{
		Width:    600,
		Height:   400,
		Format:   "avif",
		Quality:  90,
		Crop:     "smart",
		Compress: true,
		Blur:     5,
		Sharpen:  true,
		Mode:     "fill",
	})
	if err != nil {
		t.Fatal(err)
	}
	// Check that all params are present
	for _, param := range []string{"w=600", "h=400", "format=avif", "q=90", "crop=smart", "compress=true", "blur=5", "sharpen=true", "mode=fill"} {
		if !contains(result, param) {
			t.Errorf("result %q missing param %q", result, param)
		}
	}
}

func TestBuildTransformURL_ExistingQueryParams(t *testing.T) {
	result, err := buildTransformURL("https://sub.gumlet.io/img.jpg?existing=true", TransformOptions{Width: 300})
	if err != nil {
		t.Fatal(err)
	}
	if !contains(result, "existing=true") {
		t.Errorf("lost existing param in %q", result)
	}
	if !contains(result, "w=300") {
		t.Errorf("missing new param in %q", result)
	}
}

func TestBuildTransformURL_NoParams(t *testing.T) {
	result, err := buildTransformURL("https://sub.gumlet.io/img.jpg", TransformOptions{})
	if err != nil {
		t.Fatal(err)
	}
	want := "https://sub.gumlet.io/img.jpg"
	if result != want {
		t.Errorf("got %q, want %q", result, want)
	}
}

func TestBuildTransformURL_InvalidURL(t *testing.T) {
	_, err := buildTransformURL("://invalid", TransformOptions{Width: 100})
	if err == nil {
		t.Fatal("expected error for invalid URL")
	}
}

func TestHumanSize(t *testing.T) {
	tests := []struct {
		input int64
		want  string
	}{
		{500, "500 B"},
		{1024, "1.0 KB"},
		{35000, "34.2 KB"},
		{1048576, "1.0 MB"},
		{2621440, "2.5 MB"},
	}
	for _, tt := range tests {
		got := humanSize(tt.input)
		if got != tt.want {
			t.Errorf("humanSize(%d) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsStr(s, substr))
}

func containsStr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
