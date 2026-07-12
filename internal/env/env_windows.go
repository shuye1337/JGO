//go:build windows

package env

import (
	"fmt"
	"os"
	"strings"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows/registry"
)

const envRegPath = "Environment"

func setEnvVarOS(name, value string) error {
	k, err := registry.OpenKey(registry.CURRENT_USER, envRegPath, registry.SET_VALUE|registry.QUERY_VALUE)
	if err != nil {
		return fmt.Errorf("failed to open registry: %w", err)
	}
	defer k.Close()

	if err := k.SetStringValue(name, value); err != nil {
		return fmt.Errorf("failed to set %s: %w", name, err)
	}
	broadcastSettingChange()
	return nil
}

func getEnvVarOS(name string) (string, error) {
	k, err := registry.OpenKey(registry.CURRENT_USER, envRegPath, registry.QUERY_VALUE)
	if err != nil {
		v := os.Getenv(name)
		if v != "" {
			return v, nil
		}
		return "", err
	}
	defer k.Close()

	val, _, err := k.GetStringValue(name)
	if err != nil {
		if err == registry.ErrNotExist {
			v := os.Getenv(name)
			if v != "" {
				return v, nil
			}
			return "", nil
		}
		return "", err
	}
	return val, nil
}

func removeEnvVarOS(name string) error {
	k, err := registry.OpenKey(registry.CURRENT_USER, envRegPath, registry.SET_VALUE|registry.QUERY_VALUE)
	if err != nil {
		return fmt.Errorf("failed to open registry: %w", err)
	}
	defer k.Close()

	if err := k.DeleteValue(name); err != nil && err != registry.ErrNotExist {
		return fmt.Errorf("failed to delete %s: %w", name, err)
	}
	broadcastSettingChange()
	return nil
}

func addToPathOS(pathEntry string) error {
	k, err := registry.OpenKey(registry.CURRENT_USER, envRegPath, registry.SET_VALUE|registry.QUERY_VALUE)
	if err != nil {
		return fmt.Errorf("failed to open registry: %w", err)
	}
	defer k.Close()

	current, _, err := k.GetStringValue("Path")
	if err != nil && err != registry.ErrNotExist {
		return fmt.Errorf("failed to read Path: %w", err)
	}

	entries := strings.Split(current, ";")
	for _, e := range entries {
		if strings.EqualFold(strings.TrimSpace(e), pathEntry) {
			return nil
		}
	}

	var newPath string
	if current == "" {
		newPath = pathEntry
	} else {
		newPath = current + ";" + pathEntry
	}

	if err := k.SetStringValue("Path", newPath); err != nil {
		return fmt.Errorf("failed to set Path: %w", err)
	}
	broadcastSettingChange()
	return nil
}

func removeFromPathOS(pathEntry string) error {
	k, err := registry.OpenKey(registry.CURRENT_USER, envRegPath, registry.SET_VALUE|registry.QUERY_VALUE)
	if err != nil {
		return fmt.Errorf("failed to open registry: %w", err)
	}
	defer k.Close()

	current, _, err := k.GetStringValue("Path")
	if err != nil {
		if err == registry.ErrNotExist {
			return nil
		}
		return fmt.Errorf("failed to read Path: %w", err)
	}

	entries := strings.Split(current, ";")
	var kept []string
	for _, e := range entries {
		if !strings.EqualFold(strings.TrimSpace(e), pathEntry) {
			kept = append(kept, e)
		}
	}
	newPath := strings.Join(kept, ";")

	if err := k.SetStringValue("Path", newPath); err != nil {
		return fmt.Errorf("failed to set Path: %w", err)
	}
	broadcastSettingChange()
	return nil
}

func setEnvVarsOS(jdkHome string) error {
	oldJavaHome := os.Getenv("JAVA_HOME")

	if err := setEnvVarOS("JAVA_HOME", jdkHome); err != nil {
		return err
	}
	binPath := "%JAVA_HOME%\\bin"
	if err := addToPathOS(binPath); err != nil {
		return err
	}

	if oldJavaHome != "" && !strings.EqualFold(oldJavaHome, jdkHome) {
		_ = removeFromPathOS(binPath)
	}
	return nil
}

func broadcastSettingChange() {
	const (
		WM_SETTINGCHANGE = 0x001A
		SMTO_ABORTIFHUNG = 0x0002
		HWND_BROADCAST   = 0xFFFF
	)

	user32 := syscall.NewLazyDLL("user32.dll")
	sendTimeout := user32.NewProc("SendMessageTimeoutW")
	envStr := syscall.StringToUTF16Ptr("Environment")

	r, _, _ := sendTimeout.Call(
		HWND_BROADCAST,
		WM_SETTINGCHANGE,
		0,
		uintptr(unsafe.Pointer(envStr)),
		SMTO_ABORTIFHUNG,
		5000,
		0,
	)
	if r == 0 {
		fmt.Fprintf(os.Stderr, "warning: failed to broadcast environment change; changes may require a restart\n")
	}
}
