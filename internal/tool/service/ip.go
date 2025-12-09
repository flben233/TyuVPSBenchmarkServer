package service

import (
	"VPSBenchmarkBackend/internal/config"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

// IPRequest represents an IP lookup request payload.
type IPRequest struct {
	Target     string `json:"target" binding:"required"`
	DataSource string `json:"dataSource" binding:"oneof=ipinfo ip-api"`
}

var httpClient = &http.Client{Timeout: 8 * time.Second}

// IPInfo queries multiple providers for IP metadata and returns any successful responses.
func IPInfo(req *IPRequest) (map[string]interface{}, error) {
	if req == nil || req.Target == "" {
		return nil, errors.New("target is required")
	}
	if req.DataSource == "ipinfo" {
		return ipinfoSource(req.Target)
	}
	return ipapiSource(req.Target)
}

func ipapiSource(addr string) (map[string]interface{}, error) {
	cfg := config.Get()
	url := fmt.Sprintf("https://ip-api.io/api/v1/ip/%s?apikey=%s", addr, cfg.IPApiKey)
	resp, err := syncRequest(url, http.MethodGet)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ip-api status %d", resp.StatusCode)
	}

	var data map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	url = fmt.Sprintf("https://ip-api.io/api/v1/risk-score/%s?apikey=%s", addr, cfg.IPApiKey)
	resp, err = syncRequest(url, http.MethodGet)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ip-api risk-score status %d", resp.StatusCode)
	}
	var riskData map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&riskData); err != nil {
		return nil, err
	}
	data["risk"] = riskData

	return data, nil
}

func ipinfoSource(addr string) (map[string]interface{}, error) {
	url := fmt.Sprintf("https://ipinfo.io/widget/demo/%s", addr)
	resp, err := syncRequest(url, http.MethodGet)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ipinfo status %d", resp.StatusCode)
	}

	var data map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	return data, nil
}

func syncRequest(url string, method string) (*http.Response, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:140.0) Gecko/20100101 Firefox/140.0")

	return httpClient.Do(req)
}
