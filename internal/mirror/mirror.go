package mirror

import (
	"sort"
	"strings"
	"sync"

	"jgo/internal/provider"
)

type Mirror interface {
	ID() string
	DisplayName() string
	SupportedSources() []string
	ListAvailable(source, os, arch, proxy string) ([]provider.JDKAsset, error)
	TestSources(source, os, arch, proxy string) ([]provider.RequestRecord, error)
}

var (
	registry   = map[string]Mirror{}
	registryMu sync.Mutex
)

func Register(m Mirror) {
	registryMu.Lock()
	defer registryMu.Unlock()
	registry[m.ID()] = m
}

func Get(id string) (Mirror, bool) {
	registryMu.Lock()
	defer registryMu.Unlock()
	m, ok := registry[id]
	return m, ok
}

func All() []Mirror {
	registryMu.Lock()
	defer registryMu.Unlock()
	ids := make([]string, 0, len(registry))
	for id := range registry {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	out := make([]Mirror, 0, len(registry))
	for _, id := range ids {
		out = append(out, registry[id])
	}
	return out
}

func supportsSource(m Mirror, source string) bool {
	for _, s := range m.SupportedSources() {
		if s == source {
			return true
		}
	}
	return false
}

func cacheKey(cfg map[string]string) string {
	if len(cfg) == 0 {
		return ""
	}
	pairs := make([]string, 0, len(cfg))
	for k, v := range cfg {
		pairs = append(pairs, k+"="+v)
	}
	sort.Strings(pairs)
	return strings.Join(pairs, ";")
}

func Resolve(cfg map[string]string) provider.MirrorSet {
	backends := make(map[string]provider.MirrorBackend)
	for source, mirrorID := range cfg {
		m, ok := Get(mirrorID)
		if !ok {
			continue
		}
		if !supportsSource(m, source) {
			continue
		}
		backends[source] = m
	}
	return provider.MirrorSet{
		Backends: backends,
		Key:      cacheKey(cfg),
	}
}
