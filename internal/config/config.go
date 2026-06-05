package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	cliconfig "github.com/nicolasacchi/clicore/config"
)

type Project struct {
	APIKey    string `toml:"api_key"`
	Subdomain string `toml:"subdomain,omitempty"`
}

type Config struct {
	DefaultProject string              `toml:"default_project,omitempty"`
	Projects       map[string]*Project `toml:"projects,omitempty"`
}

func configFilePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "gumlet", "config.toml"), nil
}

func loadConfigFile() (*Config, error) {
	path, err := configFilePath()
	if err != nil {
		return nil, err
	}
	var cfg Config
	if _, err := toml.DecodeFile(path, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func saveConfigFile(cfg *Config) error {
	path, err := configFilePath()
	if err != nil {
		return err
	}
	// Atomic temp+rename (clicore) — replaces the O_TRUNC write so an
	// interrupted encode can't corrupt a config that already holds credentials.
	return cliconfig.SaveTOML(path, cfg)
}

func resolveProject(cfg *Config, projectFlag string) *Project {
	if cfg == nil {
		return nil
	}
	if projectFlag != "" && cfg.Projects != nil {
		if p, ok := cfg.Projects[projectFlag]; ok {
			return p
		}
		return nil
	}
	if cfg.DefaultProject != "" && cfg.Projects != nil {
		if p, ok := cfg.Projects[cfg.DefaultProject]; ok {
			return p
		}
	}
	return nil
}

// LoadAPIKey resolves the API key from flag > env > config file.
func LoadAPIKey(flagValue, projectFlag string) (string, error) {
	if flagValue != "" {
		return flagValue, nil
	}
	if v := os.Getenv("GUMLET_API_KEY"); v != "" {
		return v, nil
	}
	if v := os.Getenv("GUMLET_KEY"); v != "" {
		return v, nil
	}
	cfg, err := loadConfigFile()
	if err == nil {
		if p := resolveProject(cfg, projectFlag); p != nil && p.APIKey != "" {
			return p.APIKey, nil
		}
	}
	return "", fmt.Errorf("API key required: use --api-key flag, GUMLET_API_KEY env var, or run 'gumlet config add'")
}

// LoadSubdomain resolves the subdomain from flag > config file.
func LoadSubdomain(flagValue, projectFlag string) string {
	if flagValue != "" {
		return flagValue
	}
	cfg, err := loadConfigFile()
	if err == nil {
		if p := resolveProject(cfg, projectFlag); p != nil && p.Subdomain != "" {
			return p.Subdomain
		}
	}
	return ""
}

func AddProject(name, apiKey, subdomain string) error {
	cfg, err := loadConfigFile()
	if err != nil {
		cfg = &Config{}
	}
	if cfg.Projects == nil {
		cfg.Projects = make(map[string]*Project)
	}
	cfg.Projects[name] = &Project{
		APIKey:    apiKey,
		Subdomain: subdomain,
	}
	if cfg.DefaultProject == "" {
		cfg.DefaultProject = name
	}
	return saveConfigFile(cfg)
}

func RemoveProject(name string) error {
	cfg, err := loadConfigFile()
	if err != nil {
		return fmt.Errorf("no config file found")
	}
	if cfg.Projects == nil {
		return fmt.Errorf("project %q not found", name)
	}
	if _, ok := cfg.Projects[name]; !ok {
		return fmt.Errorf("project %q not found", name)
	}
	delete(cfg.Projects, name)
	if cfg.DefaultProject == name {
		cfg.DefaultProject = ""
		for k := range cfg.Projects {
			cfg.DefaultProject = k
			break
		}
	}
	if len(cfg.Projects) == 0 {
		cfg.Projects = nil
	}
	return saveConfigFile(cfg)
}

func SetDefaultProject(name string) error {
	cfg, err := loadConfigFile()
	if err != nil {
		return fmt.Errorf("no config file found")
	}
	if cfg.Projects == nil {
		return fmt.Errorf("project %q not found", name)
	}
	if _, ok := cfg.Projects[name]; !ok {
		return fmt.Errorf("project %q not found", name)
	}
	cfg.DefaultProject = name
	return saveConfigFile(cfg)
}

func ListProjects() (*Config, error) {
	return loadConfigFile()
}

func MaskKey(key string) string {
	if len(key) <= 10 {
		return "***"
	}
	return key[:8] + "***" + key[len(key)-4:]
}
