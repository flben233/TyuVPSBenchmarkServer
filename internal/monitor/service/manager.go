package service

import (
	"VPSBenchmarkBackend/internal/monitor/response"
	"VPSBenchmarkBackend/internal/monitor/store"
)

// 这边的方法需要区分是否为管理员，管理员可以操作所有用户的数据，普通用户只能操作自己的数据

func AddHost(username, target, name string) (int64, error) {
	return store.AddHost(target, name, username)
}

func RemoveHost(username string, id int64, isAdmin bool) error {
	if isAdmin {
		return store.RemoveHostAsAdmin(id)
	}
	return store.RemoveHost(id, username)
}

func ListHosts(username string, isAdmin bool) ([]response.HostResponse, error) {
	var hosts []store.MonitorHost
	var err error

	if isAdmin {
		hosts, err = store.ListAllHosts()
	} else {
		hosts, err = store.ListHostsByUploader(username)
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
