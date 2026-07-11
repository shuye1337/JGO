package wrapper

import (
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
)

func mavenSettingsPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".m2", "settings.xml"), nil
}

type mavenSettings struct {
	XMLName xml.Name      `xml:"settings"`
	Proxies []mavenProxy  `xml:"proxies>proxy"`
	Inner   string        `xml:",innerxml"`
}

type mavenProxy struct {
	ID       string `xml:"id"`
	Active   string `xml:"active"`
	Protocol string `xml:"protocol"`
	Host     string `xml:"host"`
	Port     string `xml:"port"`
	Username string `xml:"username,omitempty"`
	Password string `xml:"password,omitempty"`
}

func SetMavenProxy(proxyStr string) error {
	p, err := parseProxyURL(proxyStr)
	if err != nil {
		return err
	}

	settingsPath, err := mavenSettingsPath()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(settingsPath), 0755); err != nil {
		return err
	}

	var settings mavenSettings
	data, err := os.ReadFile(settingsPath)
	if err == nil {
		if err := xml.Unmarshal(data, &settings); err != nil {
			return fmt.Errorf("failed to parse existing %s: %w", settingsPath, err)
		}
	}

	if p == nil {
		settings.Proxies = nil
	} else {
		settings.Proxies = []mavenProxy{
			{
				ID:       "jgo-http",
				Active:   "true",
				Protocol: "http",
				Host:     p.host,
				Port:     p.port,
				Username: p.user,
				Password: p.password,
			},
			{
				ID:       "jgo-https",
				Active:   "true",
				Protocol: "https",
				Host:     p.host,
				Port:     p.port,
				Username: p.user,
				Password: p.password,
			},
		}
	}

	output, err := xml.MarshalIndent(settings, "", "  ")
	if err != nil {
		return err
	}
	content := xml.Header + string(output) + "\n"
	return os.WriteFile(settingsPath, []byte(content), 0644)
}

func GetMavenProxy() (string, error) {
	settingsPath, err := mavenSettingsPath()
	if err != nil {
		return "", err
	}

	data, err := os.ReadFile(settingsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "No proxy configured", nil
		}
		return "", err
	}

	var settings mavenSettings
	if err := xml.Unmarshal(data, &settings); err != nil {
		return "", err
	}

	for _, proxy := range settings.Proxies {
		if proxy.Active == "true" && proxy.Protocol == "https" {
			if proxy.Username != "" {
				return fmt.Sprintf("%s@%s:%s", proxy.Username, proxy.Host, proxy.Port), nil
			}
			return fmt.Sprintf("%s:%s", proxy.Host, proxy.Port), nil
		}
	}

	return "No proxy configured", nil
}
