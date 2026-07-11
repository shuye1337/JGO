package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type JDKInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Major   int    `json:"major"`
	Source  string `json:"source"`
	Path    string `json:"path"`
}

type Config struct {
	RootPath string             `json:"root_path"`
	Proxy    string             `json:"proxy"`
	Mirrors  map[string]string  `json:"mirrors"`
	JDKs     map[string]JDKInfo `json:"jdks"`
	Current  string             `json:"current"`
}

var GetConfigDir = func() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".jgo"), nil
}

func configPath() (string, error) {
	dir, err := GetConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.json"), nil
}

func Default() *Config {
	home, _ := os.UserHomeDir()
	return &Config{
		RootPath: filepath.Join(home, ".jgo", "jdks"),
		Proxy:    "",
		JDKs:     make(map[string]JDKInfo),
		Current:  "",
	}
}

func Load() (*Config, error) {
	p, err := configPath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			return Default(), nil
		}
		return nil, err
	}
	var c Config
	if err := json.Unmarshal(data, &c); err != nil {
		return nil, err
	}
	if c.JDKs == nil {
		c.JDKs = make(map[string]JDKInfo)
	}
	if c.Mirrors == nil {
		c.Mirrors = make(map[string]string)
	}
	return &c, nil
}

func (c *Config) Save() error {
	dir, err := GetConfigDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	p := filepath.Join(dir, "config.json")
	tmp := p + ".tmp"
	if err := os.WriteFile(tmp, data, 0644); err != nil {
		return err
	}
	return os.Rename(tmp, p)
}

func (c *Config) EnsureRoot() error {
	if c.RootPath == "" {
		return ErrRootNotSet
	}
	return os.MkdirAll(c.RootPath, 0755)
}
