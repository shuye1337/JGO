package provider

import (
	"encoding/json"
	"net/http"
	"net/url"
	"time"
)

func httpClient(proxy string) *http.Client {
	tr := &http.Transport{
		MaxIdleConns:        10,
		IdleConnTimeout:     30 * time.Second,
		DisableCompression:  false,
		MaxIdleConnsPerHost: 5,
	}
	if proxy != "" {
		if u, err := url.Parse(proxy); err == nil {
			tr.Proxy = http.ProxyURL(u)
		}
	}
	return &http.Client{
		Transport: tr,
		Timeout:   30 * time.Second,
	}
}

func fetchJSON(rawURL, proxy string, v interface{}) error {
	client := httpClient(proxy)
	req, err := http.NewRequest("GET", rawURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "jgo/1.0")
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return &fetchError{Code: resp.StatusCode, URL: rawURL}
	}
	return json.NewDecoder(resp.Body).Decode(v)
}

type fetchError struct {
	Code int
	URL  string
}

func (e *fetchError) Error() string {
	return "HTTP " + http.StatusText(e.Code) + " (" + e.URL + ")"
}

type RequestRecord struct {
	URL      string
	Duration time.Duration
	Status   string
}

func fetchJSONTimed(rawURL, proxy string, v interface{}) (RequestRecord, error) {
	start := time.Now()
	err := fetchJSON(rawURL, proxy, v)
	rec := RequestRecord{URL: rawURL, Duration: time.Since(start)}
	if err != nil {
		rec.Status = err.Error()
	} else {
		rec.Status = "OK"
	}
	return rec, err
}
