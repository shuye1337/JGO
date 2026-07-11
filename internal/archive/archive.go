package archive

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

func ValidateJDK(jdkHome string) error {
	binDir := filepath.Join(jdkHome, "bin")
	javaName := "java"
	javacName := "javac"
	if runtime.GOOS == "windows" {
		javaName = "java.exe"
		javacName = "javac.exe"
	}
	javaPath := filepath.Join(binDir, javaName)
	javacPath := filepath.Join(binDir, javacName)

	if _, err := os.Stat(javaPath); err != nil {
		return fmt.Errorf("not a valid JDK: missing %s", javaPath)
	}
	if _, err := os.Stat(javacPath); err != nil {
		return fmt.Errorf("not a valid JDK: missing %s", javacPath)
	}
	return nil
}

func Extract(archivePath, destDir string) (string, error) {
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return "", err
	}

	if strings.HasSuffix(strings.ToLower(archivePath), ".zip") {
		return extractZip(archivePath, destDir)
	}
	if strings.HasSuffix(strings.ToLower(archivePath), ".tar.gz") || strings.HasSuffix(strings.ToLower(archivePath), ".tgz") {
		return extractTarGz(archivePath, destDir)
	}
	return "", fmt.Errorf("unsupported archive format: %s", archivePath)
}

func extractZip(archivePath, destDir string) (string, error) {
	r, err := zip.OpenReader(archivePath)
	if err != nil {
		return "", fmt.Errorf("failed to open zip: %w", err)
	}
	defer r.Close()

	var topDir string
	for _, f := range r.File {
		parts := strings.Split(strings.Trim(filepath.ToSlash(f.Name), "/"), "/")
		if len(parts) > 0 && parts[0] != "" {
			topDir = parts[0]
			break
		}
	}

	for _, f := range r.File {
		name := filepath.FromSlash(f.Name)
		if err := extractZipFile(f, name, destDir); err != nil {
			return "", err
		}
	}

	if topDir != "" {
		jdkHome := filepath.Join(destDir, topDir)
		if err := ValidateJDK(jdkHome); err == nil {
			return jdkHome, nil
		}
	}

	return findJDKHome(destDir)
}

func extractZipFile(f *zip.File, name, destDir string) error {
	fullPath := filepath.Join(destDir, name)

	if f.FileInfo().IsDir() {
		return os.MkdirAll(fullPath, 0755)
	}

	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return err
	}

	rc, err := f.Open()
	if err != nil {
		return err
	}
	defer rc.Close()

	out, err := os.OpenFile(fullPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, rc)
	return err
}

func extractTarGz(archivePath, destDir string) (string, error) {
	f, err := os.Open(archivePath)
	if err != nil {
		return "", fmt.Errorf("failed to open archive: %w", err)
	}
	defer f.Close()

	gz, err := gzip.NewReader(f)
	if err != nil {
		return "", fmt.Errorf("failed to open gzip: %w", err)
	}
	defer gz.Close()

	tr := tar.NewReader(gz)

	var topDir string
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", fmt.Errorf("tar read error: %w", err)
		}
		parts := strings.Split(strings.Trim(filepath.ToSlash(hdr.Name), "/"), "/")
		if len(parts) > 0 && parts[0] != "" {
			topDir = parts[0]
			break
		}
	}

	f2, err := os.Open(archivePath)
	if err != nil {
		return "", err
	}
	defer f2.Close()

	gz2, err := gzip.NewReader(f2)
	if err != nil {
		return "", err
	}
	defer gz2.Close()

	tr2 := tar.NewReader(gz2)
	for {
		hdr, err := tr2.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", fmt.Errorf("tar read error: %w", err)
		}
		name := filepath.FromSlash(hdr.Name)
		fullPath := filepath.Join(destDir, name)

		switch hdr.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(fullPath, os.FileMode(hdr.Mode)); err != nil {
				return "", err
			}
		case tar.TypeSymlink:
			if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
				return "", err
			}
			os.Remove(fullPath)
			if err := os.Symlink(hdr.Linkname, fullPath); err != nil {
			}
		case tar.TypeReg, tar.TypeRegA:
			if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
				return "", err
			}
			out, err := os.OpenFile(fullPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.FileMode(hdr.Mode))
			if err != nil {
				return "", err
			}
			if _, err := io.Copy(out, tr2); err != nil {
				out.Close()
				return "", err
			}
			out.Close()
		}
	}

	if topDir != "" {
		jdkHome := filepath.Join(destDir, topDir)
		if err := ValidateJDK(jdkHome); err == nil {
			return jdkHome, nil
		}
	}

	return findJDKHome(destDir)
}

func findJDKHome(destDir string) (string, error) {
	entries, err := os.ReadDir(destDir)
	if err != nil {
		return "", err
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		candidate := filepath.Join(destDir, e.Name())
		if err := ValidateJDK(candidate); err == nil {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("could not find a valid JDK home in extracted archive")
}
