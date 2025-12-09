package handler

import (
	"VPSBenchmarkBackend/internal/common"
	"VPSBenchmarkBackend/internal/tool/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// IPLookup handles IP information lookup requests.
func IPLookup(ctx *gin.Context) {
	var req service.IPRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		target := ctx.Query("target")
		if target == "" {
			ctx.JSON(http.StatusBadRequest, common.Error(common.BadRequestCode, "target is required"))
			return
		}
		req.Target = target
	}

	results, err := service.IPInfo(&req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, common.Error(common.InternalErrorCode, err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, common.Success(results))
}

// Traceroute handles traceroute requests.
func Traceroute(ctx *gin.Context) {
	var req service.TracertRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		req.Target = ctx.Query("target")
		req.Mode = ctx.DefaultQuery("mode", "icmp")
		if portStr := ctx.Query("port"); portStr != "" {
			if port, parseErr := strconv.ParseUint(portStr, 10, 16); parseErr == nil {
				req.Port = uint16(port)
			}
		}
	}

	if req.Target == "" {
		ctx.JSON(http.StatusBadRequest, common.Error(common.BadRequestCode, "target is required"))
		return
	}
	if req.Mode != "icmp" && req.Mode != "tcp" {
		ctx.JSON(http.StatusBadRequest, common.Error(common.BadRequestCode, "mode must be icmp or tcp"))
		return
	}

	output := service.Traceroute(&req)
	ctx.JSON(http.StatusOK, common.Success(gin.H{"raw": output}))
}

// Whois handles WHOIS lookup requests.
func Whois(ctx *gin.Context) {
	var req service.WhoisRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		target := ctx.Query("target")
		if target == "" {
			ctx.JSON(http.StatusBadRequest, common.Error(common.BadRequestCode, "target is required"))
			return
		}
		req.Target = target
	}

	output := service.Whois(&req)
	ctx.JSON(http.StatusOK, common.Success(gin.H{"raw": output}))
}
