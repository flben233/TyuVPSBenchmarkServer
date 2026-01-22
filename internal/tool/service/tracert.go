package service

import (
	"VPSBenchmarkBackend/internal/config"
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
)

type TracertRequest struct {
	Target string `json:"target" binding:"required"`
	Mode   string `json:"mode" binding:"required,oneof=icmp tcp"`
	Port   uint16 `json:"port"`
}

func ExportTracert(req *TracertRequest) string {
	// Needs to install nexttrace
	var cmd *exec.Cmd
	params := []string{"--json"}
	if req.Mode == "tcp" {
		params = append(params, "--tcp")
		if req.Port != 0 {
			params = append(params, "--port", strconv.FormatUint(uint64(req.Port), 10), req.Target)
		} else {
			params = append(params, req.Target)
		}
	} else {
		params = append(params, req.Target)
	}
	cmd = exec.Command("nexttrace", params...)
	result, _ := cmd.Output()
	return string(result)
}

func Traceroute(req *TracertRequest) (error, map[string]interface{}) {
	reqBody, err := json.Marshal(req)
	if err != nil {
		return err, nil
	}
	resp, err := http.Post(config.Get().ExporterURL+"/tracert", "application/json", bytes.NewReader(reqBody))
	if err != nil {
		return err, nil
	}
	defer resp.Body.Close()
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return err, nil
	}
	bodyStr := string(bodyBytes)[strings.Index(string(bodyBytes), "{"):]
	var result map[string]interface{}
	err = json.Unmarshal([]byte(bodyStr), &result)
	if err != nil {
		return err, nil
	}
	return nil, result
}
