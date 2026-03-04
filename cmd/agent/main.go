package main

import (
	"VPSBenchmarkBackend/internal/inspector/service"
	"log"
	"os"
	"strconv"
	"time"
)

func main() {
	// 读环境变量
	envHostID := os.Getenv("INSPECTOR_HOST_ID")
	serverURL := os.Getenv("INSPECTOR_SERVER_URL")
	if envHostID == "" || serverURL == "" {
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
	for t := range time.Tick(service.UploadInterval) {
		service.UploadServerStatus(hostID, serverURL, iface)
		log.Printf("Uploaded server status at %v", t)
	}
}
