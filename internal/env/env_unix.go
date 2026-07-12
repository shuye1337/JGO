//go:build !windows

package env

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type shellKind int

const (
	shellBash shellKind = iota
	shellZsh
	shellFish
	shellDefault
)

func detectShell() shellKind {
	shell := os.Getenv("SHELL")
	switch {
	case strings.Contains(shell, "zsh"):
		return shellZsh
	case strings.Contains(shell, "bash"):
		return shellBash
	case strings.Contains(shell, "fish"):
		return shellFish
	default:
		return shellDefault
	}
}

func setEnvVarOS(name, value string) error {
	return updateShellProfile(name, value)
}

func getEnvVarOS(name string) (string, error) {
	return os.Getenv(name), nil
}

func removeEnvVarOS(name string) error {
	return removeFromShellProfile(name)
}

func addToPathOS(pathEntry string) error {
	pathMarker := "# jgo PATH"
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	profile := getProfilePath(home)
	if err := ensureProfileDir(profile); err != nil {
		return err
	}
	content, _ := os.ReadFile(profile)

	sh := detectShell()
	var exportLine string
	switch sh {
	case shellFish:
		exportLine = fmt.Sprintf(`set -gx PATH "%s" $PATH`, pathEntry)
	default:
		exportLine = fmt.Sprintf(`export PATH="%s:$PATH"`, pathEntry)
	}

	lines := strings.Split(string(content), "\n")
	var out []string
	found := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, pathMarker) {
			found = true
			out = append(out, pathMarker)
			out = append(out, exportLine)
			continue
		}
		if found && (strings.HasPrefix(trimmed, "export PATH=") || strings.HasPrefix(trimmed, "set -gx PATH ")) {
			continue
		}
		out = append(out, line)
	}
	if !found {
		out = append(out, pathMarker)
		out = append(out, exportLine)
	}

	return os.WriteFile(profile, []byte(strings.Join(out, "\n")), 0644)
}

func removeFromPathOS(pathEntry string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	profile := getProfilePath(home)
	if err := ensureProfileDir(profile); err != nil {
		return err
	}
	content, _ := os.ReadFile(profile)

	lines := strings.Split(string(content), "\n")
	var out []string
	skipNext := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "# jgo PATH" {
			skipNext = true
			continue
		}
		if skipNext {
			skipNext = false
			if strings.Contains(line, pathEntry) {
				continue
			}
			out = append(out, "# jgo PATH")
			out = append(out, line)
			continue
		}
		out = append(out, line)
	}

	return os.WriteFile(profile, []byte(strings.Join(out, "\n")), 0644)
}

func setEnvVarsOS(jdkHome string) error {
	oldJavaHome := os.Getenv("JAVA_HOME")

	if err := setEnvVarOS("JAVA_HOME", jdkHome); err != nil {
		return err
	}
	binPath := "$JAVA_HOME/bin"
	if err := addToPathOS(binPath); err != nil {
		return err
	}

	if oldJavaHome != "" && oldJavaHome != jdkHome {
		_ = removeFromPathOS(binPath)
	}

	home, _ := os.UserHomeDir()
	profile := getProfilePath(home)
	fmt.Fprintf(os.Stderr, "Please restart your shell or run: source %s\n", profile)
	return nil
}

func updateShellProfile(name, value string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	profile := getProfilePath(home)
	if err := ensureProfileDir(profile); err != nil {
		return err
	}
	content, _ := os.ReadFile(profile)

	sh := detectShell()
	marker := fmt.Sprintf("# jgo %s", name)
	var exportLine string
	switch sh {
	case shellFish:
		exportLine = fmt.Sprintf(`set -gx %s "%s"`, name, value)
	default:
		exportLine = fmt.Sprintf(`export %s="%s"`, name, value)
	}

	lines := strings.Split(string(content), "\n")
	var out []string
	found := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == marker {
			found = true
			out = append(out, marker)
			out = append(out, exportLine)
			continue
		}
		if found && (strings.HasPrefix(trimmed, "export "+name+"=") || strings.HasPrefix(trimmed, "set -gx "+name+" ")) {
			continue
		}
		out = append(out, line)
	}
	if !found {
		if len(out) > 0 && out[len(out)-1] != "" {
			out = append(out, "")
		}
		out = append(out, marker)
		out = append(out, exportLine)
	}

	return os.WriteFile(profile, []byte(strings.Join(out, "\n")), 0644)
}

func removeFromShellProfile(name string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	profile := getProfilePath(home)
	if err := ensureProfileDir(profile); err != nil {
		return err
	}
	content, _ := os.ReadFile(profile)

	marker := fmt.Sprintf("# jgo %s", name)
	lines := strings.Split(string(content), "\n")
	var out []string
	skipNext := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == marker {
			skipNext = true
			continue
		}
		if skipNext {
			skipNext = false
			continue
		}
		out = append(out, line)
	}

	return os.WriteFile(profile, []byte(strings.Join(out, "\n")), 0644)
}

func getProfilePath(home string) string {
	shell := os.Getenv("SHELL")
	switch {
	case strings.Contains(shell, "zsh"):
		return filepath.Join(home, ".zshrc")
	case strings.Contains(shell, "bash"):
		return filepath.Join(home, ".bashrc")
	case strings.Contains(shell, "fish"):
		return filepath.Join(home, ".config", "fish", "config.fish")
	default:
		return filepath.Join(home, ".profile")
	}
}

func ensureProfileDir(profile string) error {
	return os.MkdirAll(filepath.Dir(profile), 0755)
}
