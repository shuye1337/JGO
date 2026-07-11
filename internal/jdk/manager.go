package jdk

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"jgo/internal/archive"
	"jgo/internal/config"
	"jgo/internal/downloader"
	"jgo/internal/env"
	"jgo/internal/mirror"
	"jgo/internal/provider"
)

type Manager struct {
	Config *config.Config
}

func NewManager(cfg *config.Config) *Manager {
	return &Manager{Config: cfg}
}

func (m *Manager) ListInstalled() []config.JDKInfo {
	result := make([]config.JDKInfo, 0, len(m.Config.JDKs))
	for _, jdk := range m.Config.JDKs {
		result = append(result, jdk)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})
	return result
}

func (m *Manager) ListAvailable() ([]provider.JDKAsset, []error) {
	mirrors := mirror.Resolve(m.Config.Mirrors)
	return provider.ListAllAvailable(provider.MapOS(), provider.MapArch(), m.Config.Proxy, mirrors)
}

func (m *Manager) FindAssets(version string) []provider.JDKAsset {
	assets, _ := m.ListAvailable()
	if version == "" {
		return assets
	}
	var filtered []provider.JDKAsset
	for _, a := range assets {
		if fmt.Sprintf("%d", a.Major) == version || strings.HasPrefix(a.Version, version) {
			filtered = append(filtered, a)
		}
	}
	return filtered
}

func (m *Manager) Install(ctx context.Context, asset provider.JDKAsset, name string) error {
	if err := m.Config.EnsureRoot(); err != nil {
		return err
	}

	if name == "" {
		name = fmt.Sprintf("%s-%d", asset.Source, asset.Major)
	}

	if _, exists := m.Config.JDKs[name]; exists {
		return fmt.Errorf("JDK '%s' already exists. Remove it first or use a different name", name)
	}

	tmpDir, err := os.MkdirTemp("", "jgo-download-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)

	fmt.Fprintf(os.Stderr, "Downloading %s %s (%s)...\n", asset.Source, asset.Version, asset.FileType)
	archivePath, err := downloader.Download(ctx, asset.URL, m.Config.Proxy, tmpDir)
	if err != nil {
		return fmt.Errorf("download failed: %w", err)
	}

	installDir := filepath.Join(m.Config.RootPath, name)
	if err := os.MkdirAll(installDir, 0755); err != nil {
		return err
	}

	fmt.Fprintf(os.Stderr, "Extracting...\n")
	jdkHome, err := archive.Extract(archivePath, installDir)
	if err != nil {
		os.RemoveAll(installDir)
		return fmt.Errorf("extraction failed: %w", err)
	}

	if err := archive.ValidateJDK(jdkHome); err != nil {
		os.RemoveAll(installDir)
		return fmt.Errorf("validation failed: %w", err)
	}

	m.Config.JDKs[name] = config.JDKInfo{
		Name:    name,
		Version: asset.Version,
		Major:   asset.Major,
		Source:  asset.Source,
		Path:    jdkHome,
	}

	if err := m.Config.Save(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Fprintf(os.Stderr, "Installed: %s (%s) -> %s\n", name, asset.Version, jdkHome)
	return nil
}

func (m *Manager) AddLocal(ctx context.Context, archivePath, name string) error {
	if err := m.Config.EnsureRoot(); err != nil {
		return err
	}

	lower := strings.ToLower(archivePath)
	if !strings.HasSuffix(lower, ".zip") && !strings.HasSuffix(lower, ".tar.gz") && !strings.HasSuffix(lower, ".tgz") {
		return fmt.Errorf("file must be .zip or .tar.gz")
	}

	if _, err := os.Stat(archivePath); err != nil {
		return fmt.Errorf("file not found: %w", err)
	}

	if _, exists := m.Config.JDKs[name]; exists {
		return fmt.Errorf("JDK '%s' already exists", name)
	}

	installDir := filepath.Join(m.Config.RootPath, name)
	if err := os.MkdirAll(installDir, 0755); err != nil {
		return err
	}

	fmt.Fprintf(os.Stderr, "Extracting %s...\n", archivePath)
	jdkHome, err := archive.Extract(archivePath, installDir)
	if err != nil {
		os.RemoveAll(installDir)
		return fmt.Errorf("extraction failed: %w", err)
	}

	if err := archive.ValidateJDK(jdkHome); err != nil {
		os.RemoveAll(installDir)
		return fmt.Errorf("not a valid JDK: %w", err)
	}

	version, major := detectVersion(jdkHome)

	m.Config.JDKs[name] = config.JDKInfo{
		Name:    name,
		Version: version,
		Major:   major,
		Source:  "local",
		Path:    jdkHome,
	}

	if err := m.Config.Save(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Fprintf(os.Stderr, "Added: %s (%s) -> %s\n", name, version, jdkHome)
	return nil
}

func (m *Manager) Use(name string) error {
	jdk, exists := m.Config.JDKs[name]
	if !exists {
		return fmt.Errorf("JDK '%s' not found", name)
	}

	if err := env.SetEnvVars(jdk.Path); err != nil {
		return fmt.Errorf("failed to set environment variables: %w", err)
	}

	m.Config.Current = name
	if err := m.Config.Save(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Fprintf(os.Stderr, "Now using JDK: %s (%s)\n", name, jdk.Version)
	fmt.Fprintf(os.Stderr, "JAVA_HOME=%s\n", jdk.Path)
	return nil
}

func (m *Manager) FindByName(name string) (*config.JDKInfo, error) {
	jdk, exists := m.Config.JDKs[name]
	if !exists {
		return nil, fmt.Errorf("JDK '%s' not found", name)
	}
	return &jdk, nil
}

func (m *Manager) FindByVersion(version string) []config.JDKInfo {
	var result []config.JDKInfo
	for _, jdk := range m.Config.JDKs {
		if fmt.Sprintf("%d", jdk.Major) == version || strings.HasPrefix(jdk.Version, version) {
			result = append(result, jdk)
		}
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})
	return result
}

func (m *Manager) Remove(name string) error {
	jdk, exists := m.Config.JDKs[name]
	if !exists {
		return fmt.Errorf("JDK '%s' not found", name)
	}

	if err := os.RemoveAll(jdk.Path); err != nil {
		return fmt.Errorf("failed to remove JDK directory: %w", err)
	}

	delete(m.Config.JDKs, name)
	if m.Config.Current == name {
		m.Config.Current = ""
	}

	if err := m.Config.Save(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Fprintf(os.Stderr, "Removed: %s\n", name)
	return nil
}

func detectVersion(jdkHome string) (string, int) {
	releaseFile := filepath.Join(jdkHome, "release")
	data, err := os.ReadFile(releaseFile)
	if err != nil {
		return "unknown", 0
	}
	lines := strings.Split(string(data), "\n")
	var ver string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "JAVA_VERSION=") {
			ver = strings.Trim(strings.TrimPrefix(line, "JAVA_VERSION="), "\"")
			break
		}
		if strings.HasPrefix(line, "JAVA_VERSION_OS=") {
			ver = strings.Trim(strings.TrimPrefix(line, "JAVA_VERSION_OS="), "\"")
		}
	}
	if ver == "" {
		return "unknown", 0
	}
	major := parseMajor(ver)
	return ver, major
}

func parseMajor(ver string) int {
	ver = strings.TrimPrefix(ver, "1.")
	parts := strings.FieldsFunc(ver, func(r rune) bool {
		return r == '.' || r == '+' || r == '-' || r == '_'
	})
	if len(parts) == 0 {
		return 0
	}
	n := 0
	for _, c := range parts[0] {
		if c >= '0' && c <= '9' {
			n = n*10 + int(c-'0')
		} else {
			break
		}
	}
	return n
}
