package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfigMirrorsField(t *testing.T) {
	dir, err := os.MkdirTemp("", "jgo-test-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	orig := GetConfigDir
	GetConfigDir = func() (string, error) { return dir, nil }
	defer func() { GetConfigDir = orig }()

	c := &Config{
		RootPath: filepath.Join(dir, "jdks"),
		Proxy:    "",
		Mirrors:  map[string]string{"Temurin": "tsinghua"},
		JDKs:     make(map[string]JDKInfo),
	}
	if err := c.Save(); err != nil {
		t.Fatalf("save failed: %v", err)
	}

	loaded, err := Load()
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}
	if loaded.Mirrors["Temurin"] != "tsinghua" {
		t.Fatalf("expected Mirrors[Temurin]=tsinghua, got %v", loaded.Mirrors)
	}
}

func TestDefaultMirrorsEmpty(t *testing.T) {
	c := Default()
	if c.Mirrors != nil && len(c.Mirrors) > 0 {
		t.Fatalf("expected empty Mirrors in default, got %v", c.Mirrors)
	}
}
