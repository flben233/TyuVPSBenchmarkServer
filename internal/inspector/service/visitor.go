package service

import (
	authStore "VPSBenchmarkBackend/internal/auth/store"
	"VPSBenchmarkBackend/internal/common"
	"VPSBenchmarkBackend/internal/inspector/response"
	"VPSBenchmarkBackend/internal/inspector/store"
	"VPSBenchmarkBackend/internal/inspector/util"
	"errors"
	"fmt"
	"log"
	"strconv"

	"gorm.io/gorm"
)

func GetVisitorPage(ownerID, start, end int64, interval string) (*response.VisitorPageData, error) {
	setting, err := store.GetSettingByUserID(ownerID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, &common.InvalidParamError{Message: "visitor page is not set up"}
		}
		return nil, fmt.Errorf("failed to get user settings: %w", err)
	}
	if !setting.VisitorEnabled {
		return nil, &common.InvalidParamError{Message: "visitor page is disabled"}
	}

	allowedHostIDs, err := parseAllowedHostIDs(setting.AllowedHostIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to parse allowed host ids: %w", err)
	}

	allowedHostSet := make(map[int64]struct{}, len(allowedHostIDs))
	for _, hostID := range allowedHostIDs {
		id, err := strconv.ParseInt(hostID, 10, 64)
		if err != nil {
			return nil, err
		}
		allowedHostSet[id] = struct{}{}
	}

	hosts, err := store.ListHostsByUser(ownerID)
	if err != nil {
		return nil, err
	}

	visitorHosts := make([]response.VisitorHostData, 0, len(allowedHostSet))
	for _, host := range hosts {
		if _, ok := allowedHostSet[host.ID]; !ok {
			continue
		}
		withLossPoints := (end - start) <= 24*3600*1000000000
		rawPingPoints, err := store.QueryPingPoints(host.ID, start, end, interval, withLossPoints)
		if err != nil {
			return nil, err
		}

		pingPoints := util.ConvertToPointVO(rawPingPoints)

		recv, sent, err := store.QueryTrafficSum(host.ID, start, end)
		if err != nil {
			return nil, err
		}
		latestPing, err := store.QueryLatestPing(host.ID, start, end)
		if err != nil {
			return nil, err
		}
		lossRate, err := store.QueryLossRate(host.ID, start, end)
		if err != nil {
			return nil, err
		}

		visitorHosts = append(visitorHosts, response.VisitorHostData{
			Ping:         pingPoints,
			Loss:         lossRate,
			Sent:         sent,
			Recv:         recv,
			MonitorType:  host.MonitorType,
			Name:         host.Name,
			Tags:         host.Tags,
			LatestPing:   latestPing,
			LastUpdate:   host.LastUpdate,
			ServerStatus: host.ServerStatus,
		})
	}

	user, err := authStore.GetUserByID(ownerID)
	if err != nil {
		log.Printf("failed to get user by id %d: %v", ownerID, err)
		return nil, fmt.Errorf("failed to get user info")
	}

	return &response.VisitorPageData{
		OwnerID:   strconv.FormatInt(ownerID, 10),
		OwnerName: user.Name,
		BgURL:     setting.BgURL,
		Hosts:     visitorHosts,
	}, nil
}
