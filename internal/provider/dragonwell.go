package provider

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

const dragonwellAPIURL = "https://dragonwell-jdk.io/releases.json"

type dragonwellReleases struct {
	OSS    dragonwellSource `json:"oss"`
	GitHub dragonwellSource `json:"github"`
}

type dragonwellSource struct {
	Extended map[string]*string `json:"extended"`
	Standard map[string]*string `json:"standard"`
}

type Dragonwell struct{}

func NewDragonwell() *Dragonwell {
	return &Dragonwell{}
}

func init() {
	Register(NewDragonwell())
}

func (d *Dragonwell) Name() string {
	return "Dragonwell"
}

func dragonwellPrefixInfo(os, arch string) (prefix, fileType string, ok bool) {
	switch os + "/" + arch {
	case "linux/x64":
		return "xurl", "tar.gz", true
	case "linux/aarch64":
		return "aurl", "tar.gz", true
	case "alpine-linux/x64":
		return "apurl", "tar.gz", true
	case "windows/x64":
		return "wurl", "zip", true
	case "linux/riscv64":
		return "rurl", "tar.gz", true
	default:
		return "", "", false
	}
}

func (d *Dragonwell) ListAvailable(os, arch, proxy string) ([]JDKAsset, error) {
	prefix, fileType, ok := dragonwellPrefixInfo(os, arch)
	if !ok {
		return nil, nil
	}

	var releases dragonwellReleases
	if err := fetchJSON(dragonwellAPIURL, proxy, &releases); err != nil {
		return nil, fmt.Errorf("dragonwell: %w", err)
	}

	editions := []struct {
		name string
		data map[string]*string
	}{
		{"extended", releases.OSS.Extended},
		{"standard", releases.OSS.Standard},
	}

	var assets []JDKAsset
	for _, ed := range editions {
		assets = append(assets, d.scanEdition(ed.name, ed.data, prefix, fileType, os, arch)...)
	}
	return assets, nil
}

func (d *Dragonwell) TestSources(os, arch, proxy string) ([]RequestRecord, error) {
	var releases dragonwellReleases
	rec, err := fetchJSONTimed(dragonwellAPIURL, proxy, &releases)
	return []RequestRecord{rec}, err
}

func (d *Dragonwell) scanEdition(edition string, data map[string]*string, prefix, fileType, os, arch string) []JDKAsset {
	var majors []int
	for key := range data {
		if !strings.HasPrefix(key, "version") {
			continue
		}
		major, err := strconv.Atoi(strings.TrimPrefix(key, "version"))
		if err != nil {
			continue
		}
		majors = append(majors, major)
	}
	sort.Ints(majors)

	var assets []JDKAsset
	for _, major := range majors {
		versionVal := data["version"+strconv.Itoa(major)]
		if versionVal == nil || *versionVal == "" || *versionVal == "0" {
			continue
		}

		urlVal, exists := data[prefix+strconv.Itoa(major)]
		if !exists || urlVal == nil || *urlVal == "" {
			continue
		}

		assets = append(assets, JDKAsset{
			Source:   "dragonwell-" + edition,
			Version:  *versionVal,
			Major:    major,
			OS:       os,
			Arch:     arch,
			FileType: fileType,
			URL:      *urlVal,
			Name:     fmt.Sprintf("Dragonwell %s %s", edition, *versionVal),
		})
	}
	return assets
}
