package provider

import (
	"fmt"
	"runtime"
	"sync"
)

type JDKAsset struct {
	Source   string
	Version  string
	Major    int
	OS       string
	Arch     string
	FileType string
	URL      string
	Checksum string
	Name     string
}

type Provider interface {
	Name() string
	ListAvailable(os, arch, proxy string) ([]JDKAsset, error)
}

type MirrorBackend interface {
	ListAvailable(source, os, arch, proxy string) ([]JDKAsset, error)
	TestSources(source, os, arch, proxy string) ([]RequestRecord, error)
}

type MirrorSet struct {
	Backends map[string]MirrorBackend
	Key      string
}

type mirroredProvider struct {
	inner   Provider
	backend MirrorBackend
}

func (m *mirroredProvider) Name() string { return m.inner.Name() }

func (m *mirroredProvider) ListAvailable(os, arch, proxy string) ([]JDKAsset, error) {
	return m.backend.ListAvailable(m.inner.Name(), os, arch, proxy)
}

func (m *mirroredProvider) TestSources(os, arch, proxy string) ([]RequestRecord, error) {
	return m.backend.TestSources(m.inner.Name(), os, arch, proxy)
}

func MapOS() string {
	switch runtime.GOOS {
	case "windows":
		return "windows"
	case "linux":
		return "linux"
	case "darwin":
		return "macos"
	default:
		return runtime.GOOS
	}
}

func MapArch() string {
	switch runtime.GOARCH {
	case "amd64":
		return "x64"
	case "arm64":
		return "aarch64"
	case "386":
		return "x86"
	default:
		return runtime.GOARCH
	}
}

var registry []Provider

func Register(p Provider) {
	registry = append(registry, p)
}

func All(mirrors MirrorSet) []Provider {
	base := make([]Provider, len(registry))
	copy(base, registry)
	if len(mirrors.Backends) == 0 {
		return base
	}
	out := make([]Provider, len(base))
	for i, p := range base {
		if mb, ok := mirrors.Backends[p.Name()]; ok {
			out[i] = &mirroredProvider{inner: p, backend: mb}
		} else {
			out[i] = p
		}
	}
	return out
}

func ListAllAvailable(osName, arch, proxy string, mirrors MirrorSet) ([]JDKAsset, []error) {
	if assets, errs, ok := GetCache(osName, arch, proxy, mirrors.Key); ok {
		return assets, errs
	}

	providers := All(mirrors)
	type result struct {
		assets []JDKAsset
		err    error
	}
	results := make([]result, len(providers))
	var wg sync.WaitGroup

	for i, p := range providers {
		wg.Add(1)
		go func(idx int, prov Provider) {
			defer wg.Done()
			assets, err := prov.ListAvailable(osName, arch, proxy)
			results[idx] = result{assets: assets, err: err}
		}(i, p)
	}
	wg.Wait()

	var assets []JDKAsset
	var errs []error
	for _, r := range results {
		if r.err != nil {
			errs = append(errs, r.err)
			continue
		}
		assets = append(assets, r.assets...)
	}

	SetCache(assets, errs, osName, arch, proxy, mirrors.Key)
	return assets, errs
}

type SourceTester interface {
	TestSources(os, arch, proxy string) ([]RequestRecord, error)
}

type SourceTestResult struct {
	Source  string
	Records []RequestRecord
	Error   error
}

func TestAllSources(osName, arch, proxy string, mirrors MirrorSet) []SourceTestResult {
	providers := All(mirrors)
	results := make([]SourceTestResult, len(providers))
	var wg sync.WaitGroup

	for i, p := range providers {
		wg.Add(1)
		go func(idx int, prov Provider) {
			defer wg.Done()
			tester, ok := prov.(SourceTester)
			if !ok {
				results[idx] = SourceTestResult{
					Source: prov.Name(),
					Error:  fmt.Errorf("source testing not supported"),
				}
				return
			}
			records, err := tester.TestSources(osName, arch, proxy)
			results[idx] = SourceTestResult{
				Source:  prov.Name(),
				Records: records,
				Error:   err,
			}
		}(i, p)
	}
	wg.Wait()
	return results
}
