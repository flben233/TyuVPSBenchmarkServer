package handler

import (
	"VPSBenchmarkBackend/internal/common"
	"VPSBenchmarkBackend/internal/tool/response"
	"VPSBenchmarkBackend/internal/tool/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type GenericMap map[string]interface{}

// Traceroute handles traceroute requests.
// @Summary Traceroute
// @Description Perform a traceroute to a target. Supports query params `target`, `mode` (icmp|tcp), and `port`.
// @Tags tool
// @Accept json
// @Produce json
// @Param target query string false "Target IP or domain"
// @Param mode query string false "Mode: icmp or tcp"
// @Param port query int false "Port for TCP mode"
// @Success 200 {object} common.APIResponse[response.RawResponse]
// @Failure 400 {object} common.APIResponse[any]
// @Failure 500 {object} common.APIResponse[any]
// @Router /tool/traceroute [get]
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

	outputErr, output := service.Traceroute(&req)
	if outputErr != nil {
		ctx.JSON(http.StatusInternalServerError, common.Error(common.InternalErrorCode, outputErr.Error()))
		return
	}
	ctx.JSON(http.StatusOK, common.Success(output))
}

// Whois handles WHOIS lookup requests.
// @Summary WHOIS Lookup
// @Description Retrieve WHOIS information for a domain or IP. Accepts query `target` or JSON body.
// @Tags tool
// @Accept json
// @Produce json
// @Param target query string false "Target domain or IP"
// @Success 200 {object} common.APIResponse[response.RawResponse]
// @Failure 400 {object} common.APIResponse[any]
// @Failure 500 {object} common.APIResponse[any]
// @Router /tool/whois [get]
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
	ctx.JSON(http.StatusOK, common.Success(response.RawResponse{Raw: output}))
}
