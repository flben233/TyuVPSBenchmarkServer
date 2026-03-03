package service

import (
	authUtil "VPSBenchmarkBackend/internal/auth/util"
	"VPSBenchmarkBackend/internal/monitor/model"
	"VPSBenchmarkBackend/internal/monitor/response"
	"VPSBenchmarkBackend/internal/monitor/store"
)

// 这边的方法需要区分是否为管理员，管理员可以操作所有用户的数据，普通用户只能操作自己的数据

type HostLimitError struct{}

func (e *HostLimitError) Error() string {
	return "Host limit reached"
}

func AddHost(userID int64, username, target, name string) (int64, error) {
	if authUtil.IsAdmin(userID) {
		cnt, err := store.CountUserHosts(userID)
		if err != nil {
			return 0, err
		}
		if authUtil.CheckHostQuota(userID, cnt) {
			return 0, &HostLimitError{}
		}
	}
	return store.AddHost(target, name, username, userID)
}

func RemoveHost(userID int64, id int64) error {
	if authUtil.IsAdmin(userID) {
		return store.RemoveHostAsAdmin(id)
	}
	return store.RemoveHost(id, userID)
}

func ListHosts(userID int64) ([]response.HostResponse, error) {
	var hosts []model.MonitorHost
	var err error

	if authUtil.IsAdmin(userID) {
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
			Id:           host.Id,
			Target:       host.Target,
			Name:         host.Name,
			UploaderName: host.UploaderName,
			ReviewStatus: int(host.ReviewStatus),
		}
	}
	return result, nil
}
