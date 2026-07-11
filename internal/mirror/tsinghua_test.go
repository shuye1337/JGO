package mirror

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"jgo/internal/provider"
)

func serveHTML(html string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(html))
	}
}

func TestTsinghuaListAvailable(t *testing.T) {
	rootHTML := `<a href="21/">21/</a><a href="8/">8/</a><a href="deb/">deb/</a>`
	leaf21 := `<a href="OpenJDK21U-jdk_x64_windows_hotspot_21.0.11_10.zip">zip</a>
<a href="OpenJDK21U-jdk_x64_windows_hotspot_21.0.11_10.msi">msi</a>`
	leaf8 := `<a href="OpenJDK8U-jdk_x64_windows_hotspot_8u492b09.zip">zip</a>`

	mux := http.NewServeMux()
	mux.HandleFunc("/Adoptium/", serveHTML(rootHTML))
	mux.HandleFunc("/Adoptium/21/jdk/x64/windows/", serveHTML(leaf21))
	mux.HandleFunc("/Adoptium/8/jdk/x64/windows/", serveHTML(leaf8))
	ts := httptest.NewServer(mux)
	defer ts.Close()

	m := &tsinghuaMirror{baseURL: ts.URL + "/Adoptium"}
	assets, err := m.ListAvailable("Temurin", "windows", "x64", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(assets) != 2 {
		t.Fatalf("expected 2 assets, got %d: %+v", len(assets), assets)
	}

	var a21, a8 *provider.JDKAsset
	for i := range assets {
		if assets[i].Major == 21 {
			a21 = &assets[i]
		}
		if assets[i].Major == 8 {
			a8 = &assets[i]
		}
	}
	if a21 == nil {
		t.Fatalf("expected major 21 asset")
	}
	if a21.Version != "21.0.11_10" {
		t.Fatalf("expected version 21.0.11_10, got %s", a21.Version)
	}
	if a21.FileType != "zip" {
		t.Fatalf("expected zip, got %s", a21.FileType)
	}
	if !strings.HasSuffix(a21.URL, "/21/jdk/x64/windows/OpenJDK21U-jdk_x64_windows_hotspot_21.0.11_10.zip") {
		t.Fatalf("unexpected URL: %s", a21.URL)
	}
	if a21.Source != "Temurin" {
		t.Fatalf("expected Source Temurin, got %s", a21.Source)
	}
	if a21.OS != "windows" {
		t.Fatalf("expected OS windows, got %s", a21.OS)
	}
	if a21.Checksum != "" {
		t.Fatalf("expected empty checksum, got %s", a21.Checksum)
	}
	if a8 == nil {
		t.Fatalf("expected major 8 asset")
	}
	if a8.Version != "8u492b09" {
		t.Fatalf("expected version 8u492b09, got %s", a8.Version)
	}
}

func TestTsinghuaExcludesMSI(t *testing.T) {
	rootHTML := `<a href="21/">21/</a>`
	leaf := `<a href="OpenJDK21U-jdk_x64_windows_hotspot_21.0.11_10.msi">msi</a>
<a href="OpenJDK21U-jdk_x64_windows_hotspot_21.0.11_10.zip">zip</a>`

	mux := http.NewServeMux()
	mux.HandleFunc("/Adoptium/", serveHTML(rootHTML))
	mux.HandleFunc("/Adoptium/21/jdk/x64/windows/", serveHTML(leaf))
	ts := httptest.NewServer(mux)
	defer ts.Close()

	m := &tsinghuaMirror{baseURL: ts.URL + "/Adoptium"}
	assets, err := m.ListAvailable("Temurin", "windows", "x64", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(assets) != 1 {
		t.Fatalf("expected 1 asset (zip only), got %d", len(assets))
	}
	if assets[0].FileType != "zip" {
		t.Fatalf("expected zip, got %s", assets[0].FileType)
	}
}

func TestTsinghuaMacOSPath(t *testing.T) {
	rootHTML := `<a href="21/">21/</a>`
	leaf := `<a href="OpenJDK21U-jdk_x64_mac_hotspot_21.0.11_10.tar.gz">tgz</a>`

	mux := http.NewServeMux()
	mux.HandleFunc("/Adoptium/", serveHTML(rootHTML))
	mux.HandleFunc("/Adoptium/21/jdk/x64/mac/", serveHTML(leaf))
	ts := httptest.NewServer(mux)
	defer ts.Close()

	m := &tsinghuaMirror{baseURL: ts.URL + "/Adoptium"}
	assets, err := m.ListAvailable("Temurin", "macos", "x64", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(assets) != 1 {
		t.Fatalf("expected 1 asset, got %d", len(assets))
	}
	if assets[0].FileType != "tar.gz" {
		t.Fatalf("expected tar.gz, got %s", assets[0].FileType)
	}
	if !strings.HasSuffix(assets[0].URL, "/21/jdk/x64/mac/OpenJDK21U-jdk_x64_mac_hotspot_21.0.11_10.tar.gz") {
		t.Fatalf("unexpected URL: %s", assets[0].URL)
	}
	if assets[0].OS != "macos" {
		t.Fatalf("expected OS macos, got %s", assets[0].OS)
	}
}

func TestTsinghuaPrefersTarGzOnLinux(t *testing.T) {
	rootHTML := `<a href="21/">21/</a>`
	leaf := `<a href="OpenJDK21U-jdk_x64_linux_hotspot_21.0.11_10.zip">zip</a>
<a href="OpenJDK21U-jdk_x64_linux_hotspot_21.0.11_10.tar.gz">tgz</a>`

	mux := http.NewServeMux()
	mux.HandleFunc("/Adoptium/", serveHTML(rootHTML))
	mux.HandleFunc("/Adoptium/21/jdk/x64/linux/", serveHTML(leaf))
	ts := httptest.NewServer(mux)
	defer ts.Close()

	m := &tsinghuaMirror{baseURL: ts.URL + "/Adoptium"}
	assets, err := m.ListAvailable("Temurin", "linux", "x64", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(assets) != 1 {
		t.Fatalf("expected 1 asset, got %d", len(assets))
	}
	if assets[0].FileType != "tar.gz" {
		t.Fatalf("expected tar.gz preferred on linux, got %s", assets[0].FileType)
	}
}

func TestTsinghuaSingleMajorFailureDoesNotBlockOthers(t *testing.T) {
	rootHTML := `<a href="21/">21/</a><a href="8/">8/</a>`
	leaf8 := `<a href="OpenJDK8U-jdk_x64_windows_hotspot_8u492b09.zip">zip</a>`

	mux := http.NewServeMux()
	mux.HandleFunc("/Adoptium/", serveHTML(rootHTML))
	mux.HandleFunc("/Adoptium/21/jdk/x64/windows/", func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	})
	mux.HandleFunc("/Adoptium/8/jdk/x64/windows/", serveHTML(leaf8))
	ts := httptest.NewServer(mux)
	defer ts.Close()

	m := &tsinghuaMirror{baseURL: ts.URL + "/Adoptium"}
	assets, err := m.ListAvailable("Temurin", "windows", "x64", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(assets) != 1 {
		t.Fatalf("expected 1 asset (major 8 only), got %d", len(assets))
	}
	if assets[0].Major != 8 {
		t.Fatalf("expected major 8, got %d", assets[0].Major)
	}
}

func TestTsinghuaIDAndDisplayName(t *testing.T) {
	m := &tsinghuaMirror{}
	if m.ID() != "tsinghua" {
		t.Fatalf("expected tsinghua, got %s", m.ID())
	}
	if m.DisplayName() == "" {
		t.Fatalf("expected non-empty display name")
	}
}

func TestTsinghuaSupportedSources(t *testing.T) {
	m := &tsinghuaMirror{}
	srcs := m.SupportedSources()
	if len(srcs) != 1 || srcs[0] != "Temurin" {
		t.Fatalf("expected [Temurin], got %v", srcs)
	}
}
