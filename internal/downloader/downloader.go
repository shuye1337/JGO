package downloader

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"time"
)

func Download(ctx context.Context, rawURL, proxy, destDir string) (string, error) {
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return "", err
	}

	tr := &http.Transport{
		MaxIdleConns:        10,
		IdleConnTimeout:     30 * time.Second,
		MaxIdleConnsPerHost: 5,
	}
	if proxy != "" {
		u, err := url.Parse(proxy)
		if err != nil {
			return "", fmt.Errorf("invalid proxy URL: %w", err)
		}
		tr.Proxy = http.ProxyURL(u)
	}

	client := &http.Client{
		Transport: tr,
		Timeout:   30 * time.Minute,
	}

	req, err := http.NewRequestWithContext(ctx, "GET", rawURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "jgo/1.0")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP %d %s (%s)", resp.StatusCode, http.StatusText(resp.StatusCode), rawURL)
	}

	filename := path.Base(rawURL)
	if filename == "" || filename == "/" || filename == "." {
		filename = "download"
	}
	destPath := filepath.Join(destDir, filename)

	out, err := os.Create(destPath)
	if err != nil {
		return "", err
	}
	defer out.Close()

	total := resp.ContentLength
	var written int64
	buf := make([]byte, 32*1024)
	start := time.Now()

	for {
		select {
		case <-ctx.Done():
			os.Remove(destPath)
			return "", ctx.Err()
		default:
		}
		n, err := resp.Body.Read(buf)
		if n > 0 {
			if _, werr := out.Write(buf[:n]); werr != nil {
				os.Remove(destPath)
				return "", werr
			}
			written += int64(n)
			if total > 0 {
				pct := float64(written) / float64(total) * 100
				elapsed := time.Since(start).Seconds()
				speed := float64(written) / (1024 * 1024) / elapsed
				fmt.Fprintf(os.Stderr, "\r  %.1f%%  %d/%d MB  %.1f MB/s", pct, written/(1024*1024), total/(1024*1024), speed)
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			os.Remove(destPath)
			return "", err
		}
	}
	fmt.Fprintln(os.Stderr)

	return destPath, nil
}
