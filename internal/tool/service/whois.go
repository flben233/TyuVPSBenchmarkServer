package service

import (
	"os/exec"
)

type WhoisRequest struct {
	Target string `json:"target" binding:"required"`
}

func Whois(req *WhoisRequest) string {
	// Needs to install whois
	cmd := exec.Command("whois", req.Target)
	result, _ := cmd.Output()
	return string(result)
}
