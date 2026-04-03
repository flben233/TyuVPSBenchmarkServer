package exporter

import (
	"VPSBenchmarkBackend/internal/mq"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
	"time"

	probing "github.com/prometheus-community/pro-bing"
)

func Probe(target string, hostID int64, monitorType, replyTo, msgId string) error {
	writeErr := func(err error) error {
		_, err = mq.PublishJSONWithID(replyTo, "", PingResp{
			HostID:      hostID,
			Lat:         0,
			MonitorType: monitorType,
		}, msgId)
		return err
	}

	latency, err := measureLatency(target, monitorType)
	if err != nil {
		return writeErr(err)
	}

	_, err = mq.PublishJSONWithID(replyTo, "", PingResp{
		HostID:      hostID,
		Lat:         latency,
		MonitorType: monitorType,
	}, msgId)
	if err != nil {
		return writeErr(fmt.Errorf("failed to write probe result to Kafka: %w", err))
	}
	return nil
}

func measureLatency(target, monitorType string) (float32, error) {
	switch strings.ToLower(strings.TrimSpace(monitorType)) {
	case "", ProbePing:
		return pingLatency(target)
	case ProbeTCP:
		return tcpLatency(target)
	case ProbeHTTP:
		return httpLatency(target)
	default:
		return 0, fmt.Errorf("unsupported monitor type: %s", monitorType)
	}
}

func pingLatency(target string) (float32, error) {
	pinger, err := probing.NewPinger(target)
	if err != nil {
		return 0, fmt.Errorf("failed to create pinger for %s: %v", target, err)
	}
	pinger.SetPrivileged(true)
	pinger.Count = 3
	pinger.Timeout = 1000 * time.Millisecond
	err = pinger.Run()
	if err != nil {
		return 0, fmt.Errorf("failed to ping %s: %v", target, err)
	}
	stats := pinger.Statistics()
	if stats == nil {
		return 0, fmt.Errorf("ping statistics is nil for %s", target)
	}
	return float32(stats.AvgRtt.Milliseconds()), nil
}

func tcpLatency(target string) (float32, error) {
	return averageProbeLatency(target, func() error {
		conn, err := net.DialTimeout("tcp", target, 3*time.Second)
		if err != nil {
			return err
		}
		return conn.Close()
	})
}

func httpLatency(target string) (float32, error) {
	client := &http.Client{Timeout: 5 * time.Second}
	return averageProbeLatency(target, func() error {
		resp, err := client.Get(target)
		if err != nil {
			return err
		}
		_ = resp.Body.Close()
		return nil
	})
}

func averageProbeLatency(target string, probe func() error) (float32, error) {
	const attempts = 3
	var total float64
	var success int
	var lastErr error
	for i := 0; i < attempts; i++ {
		startedAt := time.Now()
		if err := probe(); err != nil {
			lastErr = err
			continue
		}
		total += float64(time.Since(startedAt).Milliseconds())
		success++
	}
	if success == 0 {
		return 0, fmt.Errorf("failed to probe %s: %w", target, lastErr)
	}
	return float32(total / float64(success)), nil
}

func Tracert(mode, target string, port uint64, replyTo, msgId string) error {
	// Needs to install nexttrace
	var cmd *exec.Cmd
	params := []string{"--json"}
	if mode == "tcp" {
		params = append(params, "--tcp")
		if port != 0 {
			params = append(params, "--port", strconv.FormatUint(port, 10), target)
		} else {
			params = append(params, target)
		}
	} else {
		params = append(params, target)
	}
	cmd = exec.Command("nexttrace", params...)
	body, err := cmd.Output()
	if err != nil {
		_, _ = mq.PublishJSONWithID(replyTo, "", TracertResp{}, msgId)
		return fmt.Errorf("failed to execute tracert command: %w", err)
	}
	bodyStr := string(body)[strings.Index(string(body), "{"):]
	var result map[string]interface{}
	err = json.Unmarshal([]byte(bodyStr), &result)
	if err != nil {
		_, _ = mq.PublishJSONWithID(replyTo, "", TracertResp{}, msgId)
		return fmt.Errorf("failed to execute tracert command: %w", err)
	}
	_, err = mq.PublishJSONWithID(replyTo, "", TracertResp{
		Result: result,
	}, msgId)
	if err != nil {
		return fmt.Errorf("failed to write tracert result to MQ: %w", err)
	}
	return nil
}
