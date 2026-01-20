package service

import (
	"log"
	"time"

	probing "github.com/prometheus-community/pro-bing"
)

func ExportQueryHosts(targets []string) map[string]float32 {
	// Query hosts
	resultCh := make(chan *probing.Statistics, len(targets))
	for _, target := range targets {
		go func(t string) {
			pinger, err := probing.NewPinger(t)
			if err != nil {
				log.Printf("Failed to create pinger for %s: %v", t, err)
				resultCh <- nil
				return
			}
			pinger.SetPrivileged(true)
			pinger.Count = 3
			pinger.Timeout = 1000 * time.Millisecond
			err = pinger.Run()
			if err != nil {
				log.Printf("Failed to ping %s: %v", t, err)
				resultCh <- nil
				return
			}
			resultCh <- pinger.Statistics()
		}(target)
	}
	results := make(map[string]float32)
	for i := 0; i < len(targets); i++ {
		stats := <-resultCh
		if stats != nil {
			results[stats.Addr] = float32(stats.AvgRtt.Milliseconds())
		}
	}
	return results
}
