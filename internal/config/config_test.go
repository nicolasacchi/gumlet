package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadAPIKey_FlagOverride(t *testing.T) {
	key, err := LoadAPIKey("flag-key", "")
	if err != nil {
		t.Fatal(err)
	}
	if key != "flag-key" {
		t.Errorf("got %q, want %q", key, "flag-key")
	}
}

func TestLoadAPIKey_EnvVar(t *testing.T) {
	t.Setenv("GUMLET_API_KEY", "env-key")
	key, err := LoadAPIKey("", "")
	if err != nil {
		t.Fatal(err)
	}
	if key != "env-key" {
		t.Errorf("got %q, want %q", key, "env-key")
	}
}

func TestLoadAPIKey_AltEnvVar(t *testing.T) {
	// GUMLET_API_KEY takes precedence over GUMLET_KEY, so clear any ambient one
	// (a real key in the dev shell) — otherwise this test is non-hermetic.
	t.Setenv("GUMLET_API_KEY", "")
	t.Setenv("GUMLET_KEY", "alt-key")
	key, err := LoadAPIKey("", "")
	if err != nil {
		t.Fatal(err)
	}
	if key != "alt-key" {
		t.Errorf("got %q, want %q", key, "alt-key")
	}
}

func TestLoadAPIKey_FlagOverridesEnv(t *testing.T) {
	t.Setenv("GUMLET_API_KEY", "env-key")
	key, err := LoadAPIKey("flag-key", "")
	if err != nil {
		t.Fatal(err)
	}
	if key != "flag-key" {
		t.Errorf("got %q, want %q", key, "flag-key")
	}
}

func TestLoadAPIKey_MissingKey(t *testing.T) {
	// Ensure no env var is set
	t.Setenv("GUMLET_API_KEY", "")
	t.Setenv("GUMLET_KEY", "")
	// Use a non-existent home to prevent config file loading
	t.Setenv("HOME", t.TempDir())
	_, err := LoadAPIKey("", "")
	if err == nil {
		t.Fatal("expected error for missing key")
	}
}

func TestLoadSubdomain_FlagOverride(t *testing.T) {
	sub := LoadSubdomain("my-sub", "")
	if sub != "my-sub" {
		t.Errorf("got %q, want %q", sub, "my-sub")
	}
}

func TestMaskKey(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"short", "***"},
		{"0123456789", "***"},
		{"0123456789abc", "01234567***9abc"},
		{"gumlet_abcdefghijklmnop", "gumlet_a***mnop"},
	}
	for _, tt := range tests {
		got := MaskKey(tt.input)
		if got != tt.want {
			t.Errorf("MaskKey(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestAddAndRemoveProject(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	// Create config dir
	os.MkdirAll(filepath.Join(tmpDir, ".config", "gumlet"), 0700)

	err := AddProject("test", "key123", "mysub")
	if err != nil {
		t.Fatal(err)
	}

	cfg, err := ListProjects()
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.Projects) != 1 {
		t.Fatalf("got %d projects, want 1", len(cfg.Projects))
	}
	if cfg.Projects["test"].APIKey != "key123" {
		t.Errorf("got key %q, want %q", cfg.Projects["test"].APIKey, "key123")
	}
	if cfg.Projects["test"].Subdomain != "mysub" {
		t.Errorf("got subdomain %q, want %q", cfg.Projects["test"].Subdomain, "mysub")
	}
	if cfg.DefaultProject != "test" {
		t.Errorf("got default %q, want %q", cfg.DefaultProject, "test")
	}

	err = RemoveProject("test")
	if err != nil {
		t.Fatal(err)
	}

	cfg, err = ListProjects()
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.Projects) != 0 {
		t.Fatalf("got %d projects after remove, want 0", len(cfg.Projects))
	}
}

func TestSetDefaultProject(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	os.MkdirAll(filepath.Join(tmpDir, ".config", "gumlet"), 0700)

	AddProject("first", "key1", "sub1")
	AddProject("second", "key2", "sub2")

	err := SetDefaultProject("second")
	if err != nil {
		t.Fatal(err)
	}

	cfg, err := ListProjects()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.DefaultProject != "second" {
		t.Errorf("got default %q, want %q", cfg.DefaultProject, "second")
	}
}

func TestLoadSubdomain_FromConfig(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	os.MkdirAll(filepath.Join(tmpDir, ".config", "gumlet"), 0700)

	AddProject("prod", "key1", "millefarmacie")

	sub := LoadSubdomain("", "")
	if sub != "millefarmacie" {
		t.Errorf("got %q, want %q", sub, "millefarmacie")
	}
}
