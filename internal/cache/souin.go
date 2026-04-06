package cache

import (
	"VPSBenchmarkBackend/internal/config"
	"net/http"
)

const (
	LookingGlassKey = "/looking-glass"
	IndexKey        = "/index"
	ReportBaseKey   = "/report/"
	SlideBaseKey    = "/slide/"
)

var client = &http.Client{}

func PurgeSouinCache(key string, keys ...string) error {
	request, err := http.NewRequest("PURGE", config.Get().SouinURL+"/purge/"+key, nil)
	if err != nil {
		return err
	}

	_, err = client.Do(request)
	if err != nil {
		return err
	}

	for _, k := range keys {
		request, err = http.NewRequest("PURGE", config.Get().SouinURL+"/purge/"+k, nil)
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
