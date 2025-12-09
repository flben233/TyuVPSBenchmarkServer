package service

import (
	"os/exec"
	"strconv"
)

type TracertRequest struct {
	Target string `json:"target" binding:"required"`
	Mode   string `json:"mode" binding:"required,oneof=icmp tcp"`
	Port   uint16 `json:"port"`
}

func Traceroute(req *TracertRequest) string {
	// Needs to install nexttrace
	var cmd *exec.Cmd
	params := []string{"--fast-trace", "--json"}
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
