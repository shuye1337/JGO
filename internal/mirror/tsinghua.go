package mirror

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"jgo/internal/provider"
)

const tsinghuaDefaultBaseURL = "https://mirrors.tuna.tsinghua.edu.cn/Adoptium"

func init() {
	Register(&tsinghuaMirror{baseURL: tsinghuaDefaultBaseURL})
}

type tsinghuaMirror struct {
	baseURL string
}

func (t *tsinghuaMirror) ID() string                 { return "tsinghua" }
func (t *tsinghuaMirror) DisplayName() string        { return "Tsinghua TUNA" }
func (t *tsinghuaMirror) SupportedSources() []string { return []string{"Temurin"} }

func mapAdoptiumOS(os string) string {
	if os == "macos" {
		return "mac"
	}
	return os
}

var majorDirRe = regexp.MustCompile(`<a href="(\d+)/"`)
var archiveRe = regexp.MustCompile(`<a href="([^"]+\.(?:zip|tar\.gz))"`)
var prefixRe = regexp.MustCompile(`^OpenJDK\d+U-jdk_[a-z0-9]+_[a-z0-9-]+_hotspot_`)

func (t *tsinghuaMirror) ListAvailable(source, os, arch, proxy string) ([]provider.JDKAsset, error) {
	if source != "Temurin" {
		return nil, nil
	}
	base := strings.TrimSuffix(t.baseURL, "/")

	rootPage, err := fetchPage(base+"/", proxy)
	if err != nil {
		return nil, fmt.Errorf("tsinghua: %w", err)
	}

	majors := majorDirRe.FindAllStringSubmatch(rootPage, -1)

	type result struct {
		asset *provider.JDKAsset
	}
	results := make([]result, len(majors))
	var wg sync.WaitGroup

	for i, m := range majors {
		wg.Add(1)
		go func(idx int, majorStr string) {
			defer wg.Done()
			results[idx] = result{asset: t.fetchOneMajor(majorStr, base, os, arch, proxy)}
		}(i, m[1])
	}
	wg.Wait()

	var assets []provider.JDKAsset
	for _, r := range results {
		if r.asset != nil {
			assets = append(assets, *r.asset)
		}
	}
	return assets, nil
}

func (t *tsinghuaMirror) fetchOneMajor(majorStr, base, origOS, arch, proxy string) *provider.JDKAsset {
	major, err := strconv.Atoi(majorStr)
	if err != nil {
		return nil
	}

	osPath := mapAdoptiumOS(origOS)
	leafURL := fmt.Sprintf("%s/%s/jdk/%s/%s/", base, majorStr, arch, osPath)

	page, err := fetchPage(leafURL, proxy)
	if err != nil {
		return nil
	}

	matches := archiveRe.FindAllStringSubmatch(page, -1)
	if len(matches) == 0 {
		return nil
	}

	var preferred string
	for _, m := range matches {
		name := m[1]
		if origOS == "windows" {
			if strings.HasSuffix(name, ".zip") {
				preferred = name
				break
			}
		} else {
			if strings.HasSuffix(name, ".tar.gz") {
				preferred = name
				break
			}
		}
	}
	if preferred == "" {
		preferred = matches[0][1]
	}

	fileType := "tar.gz"
	if strings.HasSuffix(preferred, ".zip") {
		fileType = "zip"
	}

	nameNoExt := preferred
	nameNoExt = strings.TrimSuffix(nameNoExt, ".zip")
	nameNoExt = strings.TrimSuffix(nameNoExt, ".tar.gz")
	version := prefixRe.ReplaceAllString(nameNoExt, "")

	return &provider.JDKAsset{
		Source:   "Temurin",
		Version:  version,
		Major:    major,
		OS:       origOS,
		Arch:     arch,
		FileType: fileType,
		URL:      leafURL + preferred,
		Checksum: "",
		Name:     "Temurin " + version,
	}
}

func (t *tsinghuaMirror) TestSources(source, os, arch, proxy string) ([]provider.RequestRecord, error) {
	if source != "Temurin" {
		return nil, nil
	}
	base := strings.TrimSuffix(t.baseURL, "/")

	start := time.Now()
	rootPage, err := fetchPage(base+"/", proxy)
	rec := provider.RequestRecord{URL: base + "/", Duration: time.Since(start)}
	if err != nil {
		rec.Status = err.Error()
		return []provider.RequestRecord{rec}, err
	}
	rec.Status = "OK"

	records := []provider.RequestRecord{rec}

	majors := majorDirRe.FindAllStringSubmatch(rootPage, -1)
	osPath := mapAdoptiumOS(os)
	for _, m := range majors {
		leafURL := fmt.Sprintf("%s/%s/jdk/%s/%s/", base, m[1], arch, osPath)
		start2 := time.Now()
		_, err := fetchPage(leafURL, proxy)
		rec2 := provider.RequestRecord{URL: leafURL, Duration: time.Since(start2)}
		if err != nil {
			rec2.Status = err.Error()
		} else {
			rec2.Status = "OK"
		}
		records = append(records, rec2)
		break
	}

	return records, nil
}

func fetchPage(rawURL, proxy string) (string, error) {
	client := newHTTPClient(proxy)
	req, err := http.NewRequest("GET", rawURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "jgo/1.0")
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP %d %s (%s)", resp.StatusCode, http.StatusText(resp.StatusCode), rawURL)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

func newHTTPClient(proxy string) *http.Client {
	tr := &http.Transport{
		MaxIdleConns:        10,
		IdleConnTimeout:     30 * time.Second,
		MaxIdleConnsPerHost: 5,
	}
	if proxy != "" {
		if u, err := url.Parse(proxy); err == nil {
			tr.Proxy = http.ProxyURL(u)
		}
	}
	return &http.Client{
		Transport: tr,
		Timeout:   30 * time.Second,
	}
}
