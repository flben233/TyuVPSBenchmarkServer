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

var lastIOCounter *psnet.IOCountersStat

const (
	UploadInterval        = 120 * time.Second
	UploadTrafficInterval = 30 * time.Minute
)

// 只有一个定时线程访问这个变量，是安全的
var counter time.Duration = 0

func uploadServerStatus(hostID int64, serverURL string, iface *string) {
	// 收集和上传数据
	status, err := perfmon.CollectServerStatus(iface)
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
	current, err := perfmon.FindNetIOCounter(iface)
	if current == nil || err != nil {
		return 0, 0
	}
	sent := current.BytesSent - lastIOCounter.BytesSent
	recv := current.BytesRecv - lastIOCounter.BytesRecv
	lastIOCounter = current
	log.Printf("Collected network traffic: iface=%s, curSent=%d, curRecv=%d, sent=%d, recv=%d", current.Name, current.BytesSent, current.BytesRecv, recv, sent)
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
	log.Printf("Using environment variables: INSPECTOR_HOST_ID=%s, INSPECTOR_SERVER_URL=%s, INSPECTOR_NETWORK_IFACE=%s", envHostID, serverURL, envIface)
	hostID, err := strconv.ParseInt(envHostID, 10, 64)
	if err != nil {
		return
	}
	lastIOCounter, err = perfmon.FindNetIOCounter(iface)
	if err != nil {
		log.Printf("Failed to find network interface %s: %v", envIface, err)
		return
	}
	log.Printf("Starting inspector agent with host ID %d, server URL %s, network iface %s", hostID, serverURL, lastIOCounter.Name)

	uploadServerStatus(hostID, serverURL, iface)
	for range time.Tick(UploadInterval) {
		uploadServerStatus(hostID, serverURL, iface)
		log.Println("Uploaded server status")
	}
}
