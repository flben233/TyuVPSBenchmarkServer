package cache

import (
	"VPSBenchmarkBackend/internal/config"
	"net/http"
)

const (
	LookingGlassKey = "/looking-glass"
	IndexKey        = "/"
	ReportBaseKey   = "/report/"
	SlideBaseKey    = "/slide/"
)

var client = &http.Client{}

func PurgeSouinCache(key string, keys ...string) error {
	souinURL := config.Get().SouinURL
	if souinURL == "" {
		return nil // No Souin URL configured, skip cache purge
	}
	request, err := http.NewRequest("PURGE", souinURL+key, nil)
	if err != nil {
		return err
	}

	_, err = client.Do(request)
	if err != nil {
		return err
	}

	for _, k := range keys {
		request, err = http.NewRequest("PURGE", souinURL+k, nil)
		if err != nil {
			return err
		}

		_, err = client.Do(request)
		if err != nil {
			return err
		}
	}
	return nil
}
