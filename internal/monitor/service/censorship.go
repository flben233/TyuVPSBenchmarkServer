package service

import (
	"VPSBenchmarkBackend/internal/common"
	"VPSBenchmarkBackend/internal/monitor/model"
	"VPSBenchmarkBackend/internal/monitor/response"
	"VPSBenchmarkBackend/internal/monitor/store"
)

// ListPendingHosts lists all hosts awaiting review (admin only)
func ListPendingHosts() ([]response.HostResponse, error) {
	hosts, err := store.ListPendingHosts()
	if err != nil {
		return nil, err
	}

	result := make([]response.HostResponse, len(hosts))
	for i, host := range hosts {
		result[i] = hostToResponse(host)
	}
	return result, nil
}

// ApproveHost approves a host for public display (admin only)
func ApproveHost(id int64) error {
	return store.UpdateReviewStatus(id, common.ReviewStatusApproved)
}

// RejectHost rejects a host (admin only)
func RejectHost(id int64) error {
	return store.UpdateReviewStatus(id, common.ReviewStatusRejected)
}

func hostToResponse(host model.MonitorHost) response.HostResponse {
	return response.HostResponse{
		Id:           host.Id,
		Target:       host.Target,
		Name:         host.Name,
		UploaderName: host.UploaderName,
		ReviewStatus: int(host.ReviewStatus),
	}
}
