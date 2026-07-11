package provider

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"runtime"
	"strconv"
	"sync"
	"time"
)

const azulBaseURL = "https://api.azul.com/metadata/v1/zulu/packages/"

type Azul struct{}

func NewAzul() *Azul {
	return &Azul{}
}

func init() {
	Register(NewAzul())
}

func (a *Azul) Name() string {
	return "Azul Zulu"
}

type azulPackage struct {
	Name            string          `json:"name"`
	JavaVersion     []int           `json:"java_version"`
	DownloadURL     string          `json:"download_url"`
	Sha256Hash      string          `json:"sha256_hash"`
	OS              string          `json:"os"`
	Arch            string          `json:"arch"`
	ArchiveType     string          `json:"archive_type"`
	JavaPackageType string          `json:"java_package_type"`
	HwBitness       json.RawMessage `json:"hw_bitness"`
}

type azulPagination struct {
	Total      int  `json:"total"`
	TotalPages int  `json:"total_pages"`
	Page       int  `json:"page"`
	NextPage   *int `json:"next_page"`
}

func (a *Azul) ListAvailable(osName, arch, proxy string) ([]JDKAsset, error) {
	type result struct {
		assets []JDKAsset
		err    error
	}
	results := make([]result, 2)
	var wg sync.WaitGroup

	for i, archiveType := range []string{"zip", "tar.gz"} {
		wg.Add(1)
		go func(idx int, at string) {
			defer wg.Done()
			list, err := a.fetchArchiveType(osName, arch, at, proxy)
			results[idx] = result{assets: list, err: err}
		}(i, archiveType)
	}
	wg.Wait()

	var assets []JDKAsset
	seen := make(map[int]bool)

	// On Unix prefer tar.gz, on Windows prefer zip
	if runtime.GOOS == "windows" {
		collectResults(results[0].assets, results[0].err, seen, &assets)
		collectResults(results[1].assets, results[1].err, seen, &assets)
	} else {
		collectResults(results[1].assets, results[1].err, seen, &assets)
		collectResults(results[0].assets, results[0].err, seen, &assets)
	}

	if len(assets) == 0 {
		for _, r := range results {
			if r.err != nil {
				return nil, r.err
			}
		}
	}
	return assets, nil
}

func (a *Azul) fetchArchiveType(osName, arch, archiveType, proxy string) ([]JDKAsset, error) {
	var assets []JDKAsset
	page := 1
	const pageSize = 1000
	for {
		params := url.Values{}
		params.Set("os", osName)
		params.Set("arch", arch)
		params.Set("java_package_type", "jdk")
		params.Set("archive_type", archiveType)
		params.Set("latest", "true")
		params.Set("release_status", "ga")
		params.Set("availability_types", "CA")
		params.Set("page", strconv.Itoa(page))
		params.Set("page_size", strconv.Itoa(pageSize))
		for _, f := range []string{
			"sha256_hash", "os", "arch", "archive_type",
			"java_package_type", "hw_bitness", "size",
		} {
			params.Add("include_fields", f)
		}
		fullURL := azulBaseURL + "?" + params.Encode()

		var pkgs []azulPackage
		pag, err := fetchAzulPage(fullURL, proxy, &pkgs)
		if err != nil {
			return nil, err
		}
		for _, p := range pkgs {
			if p.JavaPackageType != "jdk" {
				continue
			}
			major := 0
			if len(p.JavaVersion) > 0 {
				major = p.JavaVersion[0]
			}
			version := formatJavaVersion(p.JavaVersion)
			assets = append(assets, JDKAsset{
				Source:   a.Name(),
				Version:  version,
				Major:    major,
				OS:       p.OS,
				Arch:     p.Arch,
				FileType: p.ArchiveType,
				URL:      p.DownloadURL,
				Checksum: p.Sha256Hash,
				Name:     fmt.Sprintf("Azul Zulu %s", version),
			})
		}
		if pag == nil || pag.NextPage == nil {
			break
		}
		page = *pag.NextPage
	}
	return assets, nil
}

func fetchAzulPage(rawURL, proxy string, v interface{}) (*azulPagination, error) {
	client := httpClient(proxy)
	req, err := http.NewRequest("GET", rawURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "jgo/1.0")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, &fetchError{Code: resp.StatusCode, URL: rawURL}
	}
	if err := json.NewDecoder(resp.Body).Decode(v); err != nil {
		return nil, err
	}
	header := resp.Header.Get("X-Pagination")
	if header == "" {
		return nil, nil
	}
	var pag azulPagination
	if err := json.Unmarshal([]byte(header), &pag); err != nil {
		return nil, nil
	}
	return &pag, nil
}

func collectResults(list []JDKAsset, listErr error, seen map[int]bool, assets *[]JDKAsset) {
	if listErr != nil {
		return
	}
	for _, asset := range list {
		if seen[asset.Major] {
			continue
		}
		seen[asset.Major] = true
		*assets = append(*assets, asset)
	}
}

func formatJavaVersion(v []int) string {
	switch len(v) {
	case 0:
		return ""
	case 1:
		return strconv.Itoa(v[0])
	case 2:
		return fmt.Sprintf("%d.%d", v[0], v[1])
	default:
		return fmt.Sprintf("%d.%d.%d", v[0], v[1], v[2])
	}
}

func (a *Azul) TestSources(osName, arch, proxy string) ([]RequestRecord, error) {
	var allRecords []RequestRecord
	var recordsMu sync.Mutex
	var wg sync.WaitGroup

	for _, archiveType := range []string{"zip", "tar.gz"} {
		wg.Add(1)
		go func(at string) {
			defer wg.Done()
			records := a.testFetchArchiveType(osName, arch, at, proxy)
			recordsMu.Lock()
			allRecords = append(allRecords, records...)
			recordsMu.Unlock()
		}(archiveType)
	}
	wg.Wait()
	return allRecords, nil
}

func (a *Azul) testFetchArchiveType(osName, arch, archiveType, proxy string) []RequestRecord {
	var records []RequestRecord
	page := 1
	const pageSize = 1000
	for {
		params := url.Values{}
		params.Set("os", osName)
		params.Set("arch", arch)
		params.Set("java_package_type", "jdk")
		params.Set("archive_type", archiveType)
		params.Set("latest", "true")
		params.Set("release_status", "ga")
		params.Set("availability_types", "CA")
		params.Set("page", strconv.Itoa(page))
		params.Set("page_size", strconv.Itoa(pageSize))
		for _, f := range []string{
			"sha256_hash", "os", "arch", "archive_type",
			"java_package_type", "hw_bitness", "size",
		} {
			params.Add("include_fields", f)
		}
		fullURL := azulBaseURL + "?" + params.Encode()

		start := time.Now()
		var pkgs []azulPackage
		pag, err := fetchAzulPage(fullURL, proxy, &pkgs)
		rec := RequestRecord{URL: fullURL, Duration: time.Since(start)}
		if err != nil {
			rec.Status = err.Error()
		} else {
			rec.Status = "OK"
		}
		records = append(records, rec)

		if err != nil || pag == nil || pag.NextPage == nil {
			break
		}
		page = *pag.NextPage
	}
	return records
}
