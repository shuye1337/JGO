package provider

import (
	"sync"
)

type cacheEntry struct {
	assets     []JDKAsset
	errs       []error
	osName     string
	arch       string
	proxy      string
	mirrorsKey string
}

var (
	cache   *cacheEntry
	cacheMu sync.Mutex
)

func SetCache(assets []JDKAsset, errs []error, osName, arch, proxy, mirrorsKey string) {
	cacheMu.Lock()
	defer cacheMu.Unlock()
	cache = &cacheEntry{
		assets:     assets,
		errs:       errs,
		osName:     osName,
		arch:       arch,
		proxy:      proxy,
		mirrorsKey: mirrorsKey,
	}
}

func GetCache(osName, arch, proxy, mirrorsKey string) ([]JDKAsset, []error, bool) {
	cacheMu.Lock()
	defer cacheMu.Unlock()
	if cache == nil {
		return nil, nil, false
	}
	if cache.osName != osName || cache.arch != arch || cache.proxy != proxy || cache.mirrorsKey != mirrorsKey {
		return nil, nil, false
	}
	return cache.assets, cache.errs, true
}

func ClearCache() {
	cacheMu.Lock()
	defer cacheMu.Unlock()
	cache = nil
}
