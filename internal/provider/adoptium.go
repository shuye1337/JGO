package provider

import (
	"net/url"
	"strconv"
	"strings"
	"sync"
)

const adoptiumAPIBase = "https://api.adoptium.net/v3"

type adoptiumAvailableReleases struct {
	AvailableReleases   []int `json:"available_releases"`
	AvailableLTSReleases []int `json:"available_lts_releases"`
	MostRecentLTS       int   `json:"most_recent_lts"`
	MostRecentFeatureRel int  `json:"most_recent_feature_release"`
	MostRecentFeatureVer int  `json:"most_recent_feature_version"`
	TipVersion          int   `json:"tip_version"`
}

type adoptiumRelease struct {
	ReleaseName string              `json:"release_name"`
	Binaries    []adoptiumBinary    `json:"binaries"`
	VersionData adoptiumVersionData `json:"version_data"`
}

type adoptiumBinary struct {
	Architecture string      `json:"architecture"`
	OS           string      `json:"os"`
	ImageType    string      `json:"image_type"`
	Package      adoptiumPkg `json:"package"`
}

type adoptiumPkg struct {
	Name     string `json:"name"`
	Link     string `json:"link"`
	Size     int64  `json:"size"`
	Checksum string `json:"checksum"`
}

type adoptiumVersionData struct {
	Major          int    `json:"major"`
	Minor          int    `json:"minor"`
	Security       int    `json:"security"`
	Build          int    `json:"build"`
	Semver         string `json:"semver"`
	OpenJDKVersion string `json:"openjdk_version"`
}

type Adoptium struct{}

func NewAdoptium() *Adoptium {
	return &Adoptium{}
}

func init() {
	Register(NewAdoptium())
}

func (a *Adoptium) Name() string {
	return "Temurin"
}

func adoptiumMapOS(os string) string {
	if os == "macos" {
		return "mac"
	}
	return os
}

func fileTypeFromName(name string) string {
	lower := strings.ToLower(name)
	if strings.HasSuffix(lower, ".zip") {
		return "zip"
	}
	if strings.HasSuffix(lower, ".tar.gz") {
		return "tar.gz"
	}
	return ""
}

func (a *Adoptium) ListAvailable(os, arch, proxy string) ([]JDKAsset, error) {
	var avail adoptiumAvailableReleases
	if err := fetchJSON(adoptiumAPIBase+"/info/available_releases", proxy, &avail); err != nil {
		return nil, err
	}

	apiOS := adoptiumMapOS(os)

	type result struct {
		asset *JDKAsset
	}

	results := make([]result, len(avail.AvailableReleases))
	var wg sync.WaitGroup

	for i, fv := range avail.AvailableReleases {
		wg.Add(1)
		go func(idx, featureVer int) {
			defer wg.Done()
			asset := a.fetchOneVersion(featureVer, apiOS, arch, os, proxy)
			results[idx] = result{asset: asset}
		}(i, fv)
	}
	wg.Wait()

	var assets []JDKAsset
	for _, r := range results {
		if r.asset != nil {
			assets = append(assets, *r.asset)
		}
	}
	return assets, nil
}

func (a *Adoptium) TestSources(os, arch, proxy string) ([]RequestRecord, error) {
	var avail adoptiumAvailableReleases
	rec, err := fetchJSONTimed(adoptiumAPIBase+"/info/available_releases", proxy, &avail)
	if err != nil {
		return []RequestRecord{rec}, err
	}

	apiOS := adoptiumMapOS(os)

	type timedResult struct {
		rec RequestRecord
	}
	timedResults := make([]timedResult, len(avail.AvailableReleases))
	var wg sync.WaitGroup

	for i, fv := range avail.AvailableReleases {
		wg.Add(1)
		go func(idx, featureVer int) {
			defer wg.Done()
			q := url.Values{}
			q.Set("os", apiOS)
			q.Set("architecture", arch)
			q.Set("image_type", "jdk")
			q.Set("page", "0")
			q.Set("page_size", "1")
			q.Set("sort_order", "DESC")
			reqURL := adoptiumAPIBase + "/assets/feature_releases/" + strconv.Itoa(featureVer) + "/ga?" + q.Encode()

			var releases []adoptiumRelease
			r, _ := fetchJSONTimed(reqURL, proxy, &releases)
			timedResults[idx] = timedResult{rec: r}
		}(i, fv)
	}
	wg.Wait()

	records := []RequestRecord{rec}
	for _, tr := range timedResults {
		records = append(records, tr.rec)
	}
	return records, nil
}

func (a *Adoptium) fetchOneVersion(featureVer int, apiOS, arch, origOS, proxy string) *JDKAsset {
	q := url.Values{}
	q.Set("os", apiOS)
	q.Set("architecture", arch)
	q.Set("image_type", "jdk")
	q.Set("page", "0")
	q.Set("page_size", "1")
	q.Set("sort_order", "DESC")
	reqURL := adoptiumAPIBase + "/assets/feature_releases/" + strconv.Itoa(featureVer) + "/ga?" + q.Encode()

	var releases []adoptiumRelease
	if err := fetchJSON(reqURL, proxy, &releases); err != nil {
		return nil
	}
	if len(releases) == 0 {
		return nil
	}
	rel := releases[0]

	for _, b := range rel.Binaries {
		if b.OS != apiOS || b.Architecture != arch || b.ImageType != "jdk" {
			continue
		}
		if b.Package.Link == "" {
			continue
		}

		ft := fileTypeFromName(b.Package.Name)
		if ft == "" {
			if apiOS == "windows" {
				ft = "zip"
			} else {
				ft = "tar.gz"
			}
		}

		version := rel.VersionData.Semver
		if version == "" {
			version = strconv.Itoa(rel.VersionData.Major) +
				"." + strconv.Itoa(rel.VersionData.Minor) +
				"." + strconv.Itoa(rel.VersionData.Security) +
				"+" + strconv.Itoa(rel.VersionData.Build)
		}

		major := rel.VersionData.Major
		if major == 0 {
			major = featureVer
		}

		return &JDKAsset{
			Source:   a.Name(),
			Version:  version,
			Major:    major,
			OS:       origOS,
			Arch:     arch,
			FileType: ft,
			URL:      b.Package.Link,
			Checksum: b.Package.Checksum,
			Name:     "Temurin " + version,
		}
	}
	return nil
}
