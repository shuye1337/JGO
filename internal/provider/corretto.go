package provider

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strconv"
)

const correttoAPIURL = "https://downloads.corretto.aws/latest-release.json"
const correttoBaseURL = "https://corretto.aws"

type correttoFile struct {
	Checksum       string `json:"checksum"`
	ChecksumSHA256 string `json:"checksum_sha256"`
	ChecksumSHA384 string `json:"checksum_sha384"`
	Resource       string `json:"resource"`
}

type Corretto struct{}

func NewCorretto() *Corretto {
	return &Corretto{}
}

func init() {
	Register(NewCorretto())
}

func (c *Corretto) Name() string {
	return "Corretto"
}

var correttoVersionRe = regexp.MustCompile(`^/downloads/resources/([^/]+)/`)

func (c *Corretto) ListAvailable(osName, arch, proxy string) ([]JDKAsset, error) {
	client := httpClient(proxy)
	req, err := http.NewRequest("GET", correttoAPIURL, nil)
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
		return nil, &fetchError{Code: resp.StatusCode, URL: correttoAPIURL}
	}

	var topLevel map[string]json.RawMessage
	if err := json.NewDecoder(resp.Body).Decode(&topLevel); err != nil {
		return nil, err
	}

	platformRaw, ok := topLevel[osName]
	if !ok {
		return nil, nil
	}

	var platform map[string]map[string]map[string]map[string]correttoFile
	if err := json.Unmarshal(platformRaw, &platform); err != nil {
		return nil, err
	}

	pkgTypes, ok := platform[arch]
	if !ok {
		return nil, nil
	}

	versions, ok := pkgTypes["jdk"]
	if !ok {
		return nil, nil
	}

	var assets []JDKAsset
	for ver, fileTypes := range versions {
		major, err := strconv.Atoi(ver)
		if err != nil {
			continue
		}

		for ft, f := range fileTypes {
			if ft != "zip" && ft != "tar.gz" {
				continue
			}

			fullURL := correttoBaseURL + f.Resource

			fullVer := ver
			if m := correttoVersionRe.FindStringSubmatch(f.Resource); len(m) >= 2 {
				fullVer = m[1]
			}

			assets = append(assets, JDKAsset{
				Source:   c.Name(),
				Version:  fullVer,
				Major:    major,
				OS:       osName,
				Arch:     arch,
				FileType: ft,
				URL:      fullURL,
				Checksum: f.ChecksumSHA256,
				Name:     "Corretto " + fullVer,
			})
		}
	}

	return assets, nil
}

func (c *Corretto) TestSources(os, arch, proxy string) ([]RequestRecord, error) {
	var topLevel map[string]json.RawMessage
	rec, err := fetchJSONTimed(correttoAPIURL, proxy, &topLevel)
	return []RequestRecord{rec}, err
}
