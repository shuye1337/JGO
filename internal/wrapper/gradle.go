package wrapper

import (
	"bufio"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

func gradlePropertiesPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".gradle", "gradle.properties"), nil
}

type proxyInfo struct {
	host     string
	port     string
	user     string
	password string
}

func parseProxyURL(proxyStr string) (*proxyInfo, error) {
	if strings.ToLower(proxyStr) == "none" {
		return nil, nil
	}

	if !strings.Contains(proxyStr, "://") {
		proxyStr = "http://" + proxyStr
	}

	u, err := url.Parse(proxyStr)
	if err != nil {
		return nil, fmt.Errorf("invalid proxy URL: %w", err)
	}

	p := &proxyInfo{
		host: u.Hostname(),
		port: u.Port(),
	}
	if u.User != nil {
		p.user = u.User.Username()
		p.password, _ = u.User.Password()
	}
	if p.port == "" {
		p.port = "8080"
	}
	return p, nil
}

func SetGradleProxy(proxyStr string) error {
	p, err := parseProxyURL(proxyStr)
	if err != nil {
		return err
	}

	propPath, err := gradlePropertiesPath()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(propPath), 0755); err != nil {
		return err
	}

	content, _ := os.ReadFile(propPath)
	lines := strings.Split(string(content), "\n")

	if p == nil {
		return writeGradleProperties(propPath, removeGradleProxyLines(lines))
	}

	var out []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if isGradleProxyLine(trimmed) {
			continue
		}
		out = append(out, line)
	}

	if len(out) > 0 && out[len(out)-1] != "" {
		out = append(out, "")
	}

	out = append(out, "# jgo proxy settings")
	out = append(out, "systemProp.http.proxyHost="+p.host)
	out = append(out, "systemProp.http.proxyPort="+p.port)
	out = append(out, "systemProp.https.proxyHost="+p.host)
	out = append(out, "systemProp.https.proxyPort="+p.port)
	if p.user != "" {
		out = append(out, "systemProp.http.proxyUser="+p.user)
		out = append(out, "systemProp.http.proxyPassword="+p.password)
		out = append(out, "systemProp.https.proxyUser="+p.user)
		out = append(out, "systemProp.https.proxyPassword="+p.password)
	}

	return writeGradleProperties(propPath, out)
}

func writeGradleProperties(path string, lines []string) error {
	return os.WriteFile(path, []byte(strings.Join(lines, "\n")), 0644)
}

func isGradleProxyLine(line string) bool {
	prefixes := []string{
		"systemProp.http.proxyHost",
		"systemProp.http.proxyPort",
		"systemProp.http.proxyUser",
		"systemProp.http.proxyPassword",
		"systemProp.https.proxyHost",
		"systemProp.https.proxyPort",
		"systemProp.https.proxyUser",
		"systemProp.https.proxyPassword",
		"# jgo proxy settings",
	}
	for _, p := range prefixes {
		if strings.HasPrefix(line, p) {
			return true
		}
	}
	return false
}

func removeGradleProxyLines(lines []string) []string {
	var out []string
	for _, line := range lines {
		if !isGradleProxyLine(strings.TrimSpace(line)) {
			out = append(out, line)
		}
	}
	return out
}

func GetGradleProxy() (string, error) {
	propPath, err := gradlePropertiesPath()
	if err != nil {
		return "", err
	}

	f, err := os.Open(propPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "No proxy configured", nil
		}
		return "", err
	}
	defer f.Close()

	var host, port, user string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "systemProp.https.proxyHost=") {
			host = strings.TrimPrefix(line, "systemProp.https.proxyHost=")
		}
		if strings.HasPrefix(line, "systemProp.https.proxyPort=") {
			port = strings.TrimPrefix(line, "systemProp.https.proxyPort=")
		}
		if strings.HasPrefix(line, "systemProp.https.proxyUser=") {
			user = strings.TrimPrefix(line, "systemProp.https.proxyUser=")
		}
	}

	if host == "" {
		return "No proxy configured", nil
	}
	if user != "" {
		return fmt.Sprintf("%s@%s:%s", user, host, port), nil
	}
	return fmt.Sprintf("%s:%s", host, port), nil
}
