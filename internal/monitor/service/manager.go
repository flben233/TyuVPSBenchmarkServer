package service

import (
	"VPSBenchmarkBackend/internal/config"
	"VPSBenchmarkBackend/internal/monitor/model"
	"VPSBenchmarkBackend/internal/monitor/response"
	"VPSBenchmarkBackend/internal/monitor/store"
)

// 这边的方法需要区分是否为管理员，管理员可以操作所有用户的数据，普通用户只能操作自己的数据

type HostLimitError struct{}

func (e *HostLimitError) Error() string {
	return "Host limit reached"
}

func AddHost(userID, username, target, name string, isAdmin bool) (int64, error) {
	if !isAdmin {
		cnt, err := store.CountUserHosts(userID)
		if err != nil {
			return 0, err
		}
		limit := config.Get().MaxHostsPerUser
		if cnt >= int64(limit) {
			return 0, &HostLimitError{}
		}
	}
	return store.AddHost(target, name, username, userID)
}

func RemoveHost(userID string, id int64, isAdmin bool) error {
	if isAdmin {
		return store.RemoveHostAsAdmin(id)
	}
	return store.RemoveHost(id, userID)
}

func ListHosts(userID string, isAdmin bool) ([]response.HostResponse, error) {
	var hosts []model.MonitorHost
	var err error

	if isAdmin {
		hosts, err = store.ListAllHosts()
	} else {
		hosts, err = store.ListHostsByUploader(userID)
	}

	if err != nil {
		return nil, err
	}

	result := make([]response.HostResponse, len(hosts))
	for i, host := range hosts {
		result[i] = response.HostResponse{
			Id:     host.Id,
			Target: host.Target,
			Name:   host.Name,
		}
	}
	return result, nil
}
