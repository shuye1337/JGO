package mirror

import (
	"testing"

	"jgo/internal/provider"
)

type testMirror struct {
	id      string
	display string
	srcs    []string
}

func (t *testMirror) ID() string                 { return t.id }
func (t *testMirror) DisplayName() string        { return t.display }
func (t *testMirror) SupportedSources() []string { return t.srcs }
func (t *testMirror) ListAvailable(source, os, arch, proxy string) ([]provider.JDKAsset, error) {
	return nil, nil
}
func (t *testMirror) TestSources(source, os, arch, proxy string) ([]provider.RequestRecord, error) {
	return nil, nil
}

func TestRegisterAndGet(t *testing.T) {
	defer func() {
		registryMu.Lock()
		delete(registry, "test-plugin")
		registryMu.Unlock()
	}()
	m := &testMirror{id: "test-plugin", display: "Test", srcs: []string{"Temurin"}}
	Register(m)

	got, ok := Get("test-plugin")
	if !ok {
		t.Fatalf("expected to find test-plugin")
	}
	if got.ID() != "test-plugin" {
		t.Fatalf("expected test-plugin, got %s", got.ID())
	}
}

func TestGetMissing(t *testing.T) {
	if _, ok := Get("nonexistent"); ok {
		t.Fatalf("expected miss for nonexistent")
	}
}

func TestResolveValidConfig(t *testing.T) {
	defer func() {
		registryMu.Lock()
		delete(registry, "test-resolve")
		registryMu.Unlock()
	}()
	Register(&testMirror{id: "test-resolve", display: "Test", srcs: []string{"Temurin"}})

	cfg := map[string]string{"Temurin": "test-resolve"}
	ms := Resolve(cfg)
	if len(ms.Backends) != 1 {
		t.Fatalf("expected 1 backend, got %d", len(ms.Backends))
	}
	if _, ok := ms.Backends["Temurin"]; !ok {
		t.Fatalf("expected Temurin backend")
	}
	if ms.Key == "" {
		t.Fatalf("expected non-empty cache key")
	}
}

func TestResolveInvalidMirrorID(t *testing.T) {
	cfg := map[string]string{"Temurin": "nonexistent"}
	ms := Resolve(cfg)
	if len(ms.Backends) != 0 {
		t.Fatalf("expected 0 backends for invalid mirror ID")
	}
}

func TestResolveUnsupportedSource(t *testing.T) {
	defer func() {
		registryMu.Lock()
		delete(registry, "test-unsupported")
		registryMu.Unlock()
	}()
	Register(&testMirror{id: "test-unsupported", display: "Test", srcs: []string{"Corretto"}})

	cfg := map[string]string{"Temurin": "test-unsupported"}
	ms := Resolve(cfg)
	if len(ms.Backends) != 0 {
		t.Fatalf("expected 0 backends for unsupported source")
	}
}

func TestResolveEmptyConfig(t *testing.T) {
	ms := Resolve(nil)
	if len(ms.Backends) != 0 {
		t.Fatalf("expected 0 backends for nil config")
	}
	if ms.Key != "" {
		t.Fatalf("expected empty key for nil config")
	}
}
