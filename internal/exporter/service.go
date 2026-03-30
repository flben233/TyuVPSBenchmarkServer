package exporter

import (
	"VPSBenchmarkBackend/internal/mq"
	"encoding/json"
	"fmt"
	"github.com/segmentio/kafka-go"
	"log"
	"os/exec"
	"strconv"
	"strings"
	"time"

	probing "github.com/prometheus-community/pro-bing"
)

var (
	pingWriter    *kafka.Writer
	tracertWriter *kafka.Writer
)

func init() {
	writer, err := mq.NewWriter(PingSentTopic)
	if err != nil {
		log.Fatalf("Failed to create Kafka writer for ping topic: %v", err)
	}
	pingWriter = writer

	writer, err = mq.NewWriter(TracertSentTopic)
	if err != nil {
		log.Fatalf("Failed to create Kafka writer for tracert topic: %v", err)
	}
	tracertWriter = writer
}

func Ping(target string, hostID int64) error {
	writeErr := func(err error) error {
		_, _ = mq.WriteJSONMessages(pingWriter, PingResp{
			HostID: hostID,
			Lat:    0,
		})
		return err
	}
	// Query hosts
	pinger, err := probing.NewPinger(target)
	if err != nil {
		return writeErr(fmt.Errorf("failed to create pinger for %s: %v", target, err))
	}
	pinger.SetPrivileged(true)
	pinger.Count = 3
	pinger.Timeout = 1000 * time.Millisecond
	err = pinger.Run()
	if err != nil {
		return writeErr(fmt.Errorf("failed to ping %s: %v", target, err))
	}
	stats := pinger.Statistics()
	if stats != nil {
		_, err = mq.WriteJSONMessages(pingWriter, PingResp{
			HostID: hostID,
			Lat:    float32(stats.AvgRtt.Milliseconds()),
		})
		if err != nil {
			return writeErr(fmt.Errorf("failed to write ping result to Kafka: %w", err))
		}
	}
	return writeErr(fmt.Errorf("ping statistics is nil for %s", target))
}

func Tracert(mode, target string, port uint64) error {
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
		_, _ = mq.WriteJSONMessages(tracertWriter, TracertResp{})
		return fmt.Errorf("failed to execute tracert command: %w", err)
	}
	bodyStr := string(body)[strings.Index(string(body), "{"):]
	var result map[string]interface{}
	err = json.Unmarshal([]byte(bodyStr), &result)
	if err != nil {
		_, _ = mq.WriteJSONMessages(tracertWriter, TracertResp{})
		return fmt.Errorf("failed to execute tracert command: %w", err)
	}
	_, err = mq.WriteJSONMessages(tracertWriter, TracertResp{
		Result: result,
	})
	if err != nil {
		return fmt.Errorf("failed to write tracert result to Kafka: %w", err)
	}
	return nil
}
