package provider

import (
	"testing"
)

type fakeBackend struct {
	assets []JDKAsset
	err    error
}

func (f *fakeBackend) ListAvailable(source, os, arch, proxy string) ([]JDKAsset, error) {
	return f.assets, f.err
}

func (f *fakeBackend) TestSources(source, os, arch, proxy string) ([]RequestRecord, error) {
	return nil, nil
}

func TestMirroredProviderDelegatesListAvailable(t *testing.T) {
	inner := NewAdoptium()
	backend := &fakeBackend{assets: []JDKAsset{{Source: "Temurin", Version: "21.0.1", Major: 21}}}
	mp := &mirroredProvider{inner: inner, backend: backend}

	assets, err := mp.ListAvailable("windows", "x64", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(assets) != 1 || assets[0].Version != "21.0.1" {
		t.Fatalf("unexpected assets: %+v", assets)
	}
}

func TestMirroredProviderPreservesName(t *testing.T) {
	inner := NewAdoptium()
	backend := &fakeBackend{}
	mp := &mirroredProvider{inner: inner, backend: backend}
	if mp.Name() != "Temurin" {
		t.Fatalf("expected Temurin, got %s", mp.Name())
	}
}

func TestAllWithEmptyMirrorsReturnsUnwrapped(t *testing.T) {
	ClearCache()
	providers := All(MirrorSet{})
	if len(providers) != 4 {
		t.Fatalf("expected 4 providers, got %d", len(providers))
	}
	for _, p := range providers {
		if _, ok := p.(*mirroredProvider); ok {
			t.Fatalf("provider %s should not be wrapped", p.Name())
		}
	}
}

func TestAllWithMirrorsWrapsMatching(t *testing.T) {
	ClearCache()
	backend := &fakeBackend{}
	mirrors := MirrorSet{
		Backends: map[string]MirrorBackend{"Temurin": backend},
		Key:      "Temurin=test",
	}
	providers := All(mirrors)
	var found bool
	for _, p := range providers {
		if p.Name() == "Temurin" {
			if _, ok := p.(*mirroredProvider); !ok {
				t.Fatalf("Temurin should be wrapped")
			}
			found = true
		}
	}
	if !found {
		t.Fatalf("Temurin provider not found")
	}
}

func TestCacheKeyIsolation(t *testing.T) {
	ClearCache()
	assets1 := []JDKAsset{{Source: "Temurin", Major: 21}}
	SetCache(assets1, nil, "windows", "x64", "", "key1")
	if got, _, ok := GetCache("windows", "x64", "", "key1"); !ok || len(got) != 1 {
		t.Fatalf("expected cache hit for key1")
	}
	if _, _, ok := GetCache("windows", "x64", "", "key2"); ok {
		t.Fatalf("expected cache miss for key2")
	}
}
