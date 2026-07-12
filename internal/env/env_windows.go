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

// readPathRaw reads the raw (unexpanded) Path value and its type from the registry.
// GetStringValue would expand %VAR% references in REG_EXPAND_SZ values, losing the
// original literal form. GetValue returns raw UTF-16LE bytes which we decode ourselves.
func readPathRaw(k registry.Key) (string, uint32, error) {
	n, valtype, err := k.GetValue("Path", nil)
	if err != nil {
		if err == registry.ErrNotExist {
			return "", registry.NONE, nil
		}
		return "", 0, err
	}
	if n == 0 {
		return "", valtype, nil
	}
	buf := make([]byte, n)
	if _, _, err = k.GetValue("Path", buf); err != nil {
		return "", 0, err
	}
	u16 := make([]uint16, 0, len(buf)/2)
	for i := 0; i+1 < len(buf); i += 2 {
		c := uint16(buf[i]) | uint16(buf[i+1])<<8
		if c == 0 {
			break
		}
		u16 = append(u16, c)
	}
	return syscall.UTF16ToString(u16), valtype, nil
}

func addToPathOS(pathEntry string) error {
	k, err := registry.OpenKey(registry.CURRENT_USER, envRegPath, registry.SET_VALUE|registry.QUERY_VALUE)
	if err != nil {
		return fmt.Errorf("failed to open registry: %w", err)
	}
	defer k.Close()

	current, _, err := readPathRaw(k)
	if err != nil {
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

	if err := k.SetExpandStringValue("Path", newPath); err != nil {
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

	current, _, err := readPathRaw(k)
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

	if err := k.SetExpandStringValue("Path", newPath); err != nil {
		return fmt.Errorf("failed to set Path: %w", err)
	}
	broadcastSettingChange()
	return nil
}

func setEnvVarsOS(jdkHome string) error {
	if err := setEnvVarOS("JAVA_HOME", jdkHome); err != nil {
		return err
	}
	binPath := "%JAVA_HOME%\\bin"
	if err := addToPathOS(binPath); err != nil {
		return err
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
