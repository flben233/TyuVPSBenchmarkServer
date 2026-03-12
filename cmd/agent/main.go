package main

import (
	"VPSBenchmarkBackend/internal/inspector/model"
	"VPSBenchmarkBackend/internal/inspector/request"
	"VPSBenchmarkBackend/pkg/perfmon"
	"bytes"
	"encoding/json"
	psnet "github.com/shirou/gopsutil/v4/net"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"
)

var lastIOCounter psnet.IOCountersStat

const (
	UploadInterval        = 120 * time.Second
	UploadTrafficInterval = 30 * time.Minute
)

// 只有一个定时线程访问这个变量，是安全的
var counter time.Duration = 0

func UploadServerStatus(hostID int64, serverURL string, iface *string) {
	// 收集和上传数据
	status, err := perfmon.CollectServerStatus()
	if err != nil {
		return
	}
	traffic := make([]model.TrafficPoint, 0)
	if counter == 0 {
		recv, sent := collectNetworkTraffic(iface)
		traffic = append(traffic, model.TrafficPoint{
			HostID: hostID,
			Recv:   float32(recv) / 1000000.0,
			Sent:   float32(sent) / 1000000.0,
			Time:   time.Now(),
		})
		counter = UploadTrafficInterval
	} else {
		counter -= UploadInterval
	}

	req := request.PutDataRequest{
		HostID:   hostID,
		HostInfo: status,
		Traffic:  traffic,
	}
	path, err := url.JoinPath(serverURL, "/api/inspector/data/put")
	if err != nil {
		return
	}
	body, err := json.Marshal(&req)
	if err != nil {
		return
	}
	_, _ = http.Post(path, "application/json", bytes.NewReader(body))
}

func collectNetworkTraffic(iface *string) (uint64, uint64) {
	counters, err := psnet.IOCounters(false)
	if err != nil || len(counters) == 0 {
		return 0, 0
	}
	current := counters[0]
	for i := 0; i < len(counters) && iface != nil; i++ {
		if counters[i].Name == *iface {
			current = counters[i]
			break
		}
	}
	sent := current.BytesSent - lastIOCounter.BytesSent
	recv := current.BytesRecv - lastIOCounter.BytesRecv
	lastIOCounter = current
	return recv, sent
}

func main() {
	// 读环境变量
	envHostID := os.Getenv("INSPECTOR_HOST_ID")
	serverURL := os.Getenv("INSPECTOR_SERVER_URL")
	if envHostID == "" || serverURL == "" {
		log.Println("Environment variables INSPECTOR_HOST_ID and INSPECTOR_SERVER_URL must be set")
		return
	}
	envIface := os.Getenv("INSPECTOR_NETWORK_IFACE")
	iface := &envIface
	if envIface == "" {
		iface = nil
	}
	hostID, err := strconv.ParseInt(envHostID, 10, 64)
	if err != nil {
		return
	}
	UploadServerStatus(hostID, serverURL, iface)
	for range time.Tick(UploadInterval) {
		UploadServerStatus(hostID, serverURL, iface)
		log.Println("Uploaded server status")
	}
}
