package service

import (
	"VPSBenchmarkBackend/internal/auth/util"
	"VPSBenchmarkBackend/internal/common"
	"VPSBenchmarkBackend/internal/config"
	"VPSBenchmarkBackend/internal/inspector/model"
	"VPSBenchmarkBackend/internal/inspector/response"
	"VPSBenchmarkBackend/internal/inspector/store"
	"VPSBenchmarkBackend/pkg/batch"
	"VPSBenchmarkBackend/pkg/perfmon"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"gorm.io/gorm"
	"io"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"
)

var putRecord = make(map[int64]time.Time)

const (
	pingInterval = 120 * time.Second
	putInterval  = 30 * time.Second
	putMaxLength = 1
)

func init() {
	common.RegisterCronJob(pingInterval, pingHosts)
}

func pingHosts() {
	hosts, err := store.ListAllHost()
	if err != nil {
		log.Printf("Failed to list all hosts for pinging: %v", err)
		return
	}
	for _, host := range hosts {
		lat, err := queryHost(host.Target)
		if err != nil {
			log.Printf("failed to query host %s: %s", host.Target, err.Error())
		}
		pingData := []model.PingPoint{{
			HostID:  host.ID,
			Latency: lat,
			Time:    time.Now(),
		}}

		if host.Notify {
			setting, err := store.GetSettingByUserID(host.UserID)
			if err != nil {
				log.Printf("Failed to get setting for user %d: %v", host.UserID, err)
			}
			if setting.NotifyURL == nil {
				return
			}
			if lat == 0 {
				points, err := store.QueryLatestNPingPoints(host.ID, host.NotifyTolerance)
				if err != nil {
					log.Printf("Failed to query latest ping points for host %d: %v", host.ID, err)
					continue
				}
				// 如果最新的 NotifyTolerance 条数据都是 0，才通知用户主机离线，避免偶尔的网络波动导致误报
				if int64(len(points)) == host.NotifyTolerance && batch.IsAllTrue(points, func(p model.PingPoint) bool { return p.Latency == 0 }) {
					tryNotify(*setting.NotifyURL, fmt.Sprintf("主机 %s (%s) 离线", host.Name, host.Target))
				}
			} else {
				points, err := store.QueryLatestNPingPoints(host.ID, host.NotifyTolerance*2)
				if err != nil {
					log.Printf("Failed to query latest ping for host %d: %v", host.ID, err)
					continue
				}
				// 如果之前的 NotifyTolerance 条数据都是 0，且最近的 NotifyTolerance 条数据不是 0，才通知用户主机上线，避免偶尔的网络波动导致误报
				if int64(len(points)) == host.NotifyTolerance*2 &&
					batch.IsAllTrue(points[:host.NotifyTolerance], func(p model.PingPoint) bool { return p.Latency == 0 }) &&
					!batch.IsAllTrue(points[int(host.NotifyTolerance):], func(p model.PingPoint) bool { return p.Latency > 0 }) {
					tryNotify(*setting.NotifyURL, fmt.Sprintf("主机 %s (%s) 上线", host.Name, host.Target))
				}
			}
		}

		if err := store.SavePingPoints(pingData); err != nil {
			log.Printf("Failed to save ping points for host %d: %v", host.ID, err)
		}
	}
}

func CreateHost(userID int64, target, name, tags string, notify bool, notifyTolerance int64) (int64, error) {
	hosts, err := store.CountUserHosts(userID)
	if err != nil {
		return 0, err
	}
	if !util.CheckInspectorQuota(userID, hosts) {
		return 0, &common.LimitExceededError{Message: fmt.Sprintf("Host limit reached: %d hosts", hosts)}
	}
	id := rand.Int63()
	for _, err = store.GetHostByID(id); !errors.Is(err, gorm.ErrRecordNotFound); _, err = store.GetHostByID(id) {
		if err != nil {
			return 0, fmt.Errorf("failed to check host ID %d: %w", id, err)
		}
		id = rand.Int63()
	}
	host := &model.InspectHost{
		ID:              id,
		UserID:          userID,
		Target:          target,
		Name:            name,
		Tags:            tags,
		Notify:          notify,
		NotifyTolerance: notifyTolerance,
	}
	if err := store.CreateHost(host); err != nil {
		return 0, err
	}
	return host.ID, nil
}

func UpdateHost(userID int64, hostID int64, name, tags, target string, notify bool, notifyTolerance int64) error {
	// 校验主机属于当前用户
	ids := store.GetHostIDByUser(userID)
	found := false
	for _, id := range ids {
		if id == hostID {
			found = true
			break
		}
	}
	if !found {
		return &common.InvalidParamError{Message: fmt.Sprintf("host %d not found or not owned by user", hostID)}
	}

	host, err := store.GetHostByID(hostID)
	if err != nil {
		return fmt.Errorf("failed to get host by ID %d: %w", hostID, err)
	}
	host.Name = name
	host.Tags = tags
	host.Target = target
	host.Notify = notify
	host.NotifyTolerance = notifyTolerance

	store.UpdateHost(host)
	return nil
}

func DeleteHost(userID int64, hostID int64) error {
	// 校验主机属于当前用户
	ids := store.GetHostIDByUser(userID)
	found := false
	for _, id := range ids {
		if id == hostID {
			found = true
			break
		}
	}
	if !found {
		return &common.InvalidParamError{Message: fmt.Sprintf("host %d not found or not owned by user", hostID)}
	}
	return store.DeleteHost(hostID)
}

func ListHosts(userID int64) ([]response.HostListResponse, error) {
	hosts, err := store.ListHostsByUser(userID)
	if err != nil {
		return nil, err
	}
	inspectHosts := make([]response.HostListResponse, len(hosts))
	for i, host := range hosts {
		inspectHosts[i] = response.HostListResponse{
			ID:           strconv.FormatInt(host.ID, 10),
			UserID:       host.UserID,
			Target:       host.Target,
			Name:         host.Name,
			Tags:         host.Tags,
			Notify:       host.Notify,
			LastUpdate:   host.LastUpdate,
			ServerStatus: host.ServerStatus,
		}
	}
	return inspectHosts, nil
}

func PutData(trafficData []model.TrafficPoint, hostInfo perfmon.ServerStatus, hostID int64) error {
	// 校验数据，理论上ID都是一样的
	_, err := store.GetHostByID(hostID)
	if err != nil {
		return &common.InvalidParamError{Message: fmt.Sprintf("host %d not found", hostID)}
	}
	for _, point := range trafficData {
		if point.HostID != hostID {
			return &common.InvalidParamError{Message: fmt.Sprintf("all data points must have the same host ID %d", hostID)}
		}
	}
	if len(trafficData) > putMaxLength {
		return &common.InvalidParamError{Message: fmt.Sprintf("too many data points, max length is %d", putMaxLength)}
	}
	if lastPut, ok := putRecord[hostID]; ok && time.Since(lastPut) < putInterval {
		return &common.RateLimitExceededError{Message: fmt.Sprintf("put data too frequently, please wait %d seconds", int(putInterval.Seconds()))}
	}
	putRecord[hostID] = time.Now()

	host, err := store.GetHostByID(hostID)
	if err != nil {
		return fmt.Errorf("failed to get host by ID %d: %w", hostID, err)
	}

	host.ServerStatus = hostInfo
	host.LastUpdate = time.Now()

	store.UpdateHost(host)

	// 保存数据
	if err := store.SaveTrafficPoints(trafficData); err != nil {
		return err
	}
	return nil
}

func QueryData(userID int64, start, end int64, interval string) ([]*response.HostData, error) {
	hosts, err := store.ListHostsByUser(userID)
	if err != nil {
		return nil, err
	}
	data := make([]*response.HostData, len(hosts))
	for i, host := range hosts {
		pingPoints, err := store.QueryPingPoints(host.ID, start, end, interval)
		if err != nil {
			return nil, err
		}
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
		data[i] = &response.HostData{
			Ping:         pingPoints,
			Loss:         lossRate,
			Sent:         sent,
			Recv:         recv,
			ID:           strconv.FormatInt(host.ID, 10),
			Target:       host.Target,
			Name:         host.Name,
			Tags:         host.Tags,
			Notify:       host.Notify,
			LatestPing:   latestPing,
			LastUpdate:   host.LastUpdate,
			ServerStatus: host.ServerStatus,
		}
	}
	return data, nil
}

func GetUserSettings(userID int64) (*response.SettingData, error) {
	setting, err := store.GetSettingByUserID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			setting = &model.InspectorSetting{
				UserID: userID,
			}
			if err := store.UpsertSetting(setting); err != nil {
				log.Printf("Failed to create default setting for user %d: %v", userID, err)
			}
		} else {
			log.Printf("Failed to get setting for user %d: %v", userID, err)
			setting = &model.InspectorSetting{UserID: userID}
		}
	}
	allowedHostIDs, err := parseAllowedHostIDs(setting.AllowedHostIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to parse allowed host IDs for user %d: %w", userID, err)
	}
	return &response.SettingData{
		NotifyURL:      setting.NotifyURL,
		BgURL:          setting.BgURL,
		VisitorEnabled: setting.VisitorEnabled,
		AllowedHostIDs: allowedHostIDs,
	}, nil
}

func UpdateUserSettings(userID int64, notifyURL, bgURL *string, visitorEnabled bool, allowedHostIDs []string) error {
	setting, err := store.GetSettingByUserID(userID)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		setting = &model.InspectorSetting{UserID: userID}
	}
	setting.BgURL = bgURL
	setting.NotifyURL = notifyURL
	setting.VisitorEnabled = visitorEnabled
	allowedHostIDsStr, err := formatAllowedHostIDs(allowedHostIDs)
	if err != nil {
		return fmt.Errorf("failed to format allowed host IDs for user %d: %w", userID, err)
	}
	setting.AllowedHostIDs = allowedHostIDsStr
	return store.UpsertSetting(setting)
}

func TestNotify(notifyURL string) {
	tryNotify(notifyURL, "这是一条测试通知")
}

func queryHost(target string) (float32, error) {
	req, err := json.Marshal([]string{target})
	if err != nil {
		log.Printf("Failed to marshal targets: %v", err)
		return 0, err
	}
	resp, err := http.Post(config.Get().ExporterURL+"/monitor", "application/json", bytes.NewReader(req))
	if err != nil {
		log.Printf("Failed to get exporter data: %v", err)
		return 0, err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Failed to read exporter response body: %v", err)
		return 0, err
	}
	var data map[string]float32
	err = json.Unmarshal(bodyBytes, &data)
	if err != nil {
		log.Printf("Failed to unmarshal exporter response: %v", err)
		return 0, err
	}
	return data[target], nil
}

func tryNotify(notifyURL string, message string) {
	// 发送通知
	go func() {
		notifyReq := map[string]interface{}{
			"urls":  notifyURL,
			"body":  message,
			"title": "Lolicon Monitor 通知",
		}
		reqBytes, err := json.Marshal(notifyReq)
		if err != nil {
			log.Printf("Failed to marshal notify request: %v", err)
			return
		}
		resp, err := http.Post(config.Get().AppriseURL, "application/json", bytes.NewReader(reqBytes))
		if err != nil {
			log.Printf("Failed to send notify request: %v", err)
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			bodyBytes, _ := io.ReadAll(resp.Body)
			log.Printf("Failed to send notify request, status: %d, response: %s", resp.StatusCode, string(bodyBytes))
		}
	}()
}

func parseAllowedHostIDs(raw string) ([]string, error) {
	if raw == "" {
		return []string{}, nil
	}

	var hostIDs []string
	if err := json.Unmarshal([]byte(raw), &hostIDs); err != nil {
		return nil, err
	}
	return hostIDs, nil
}

func formatAllowedHostIDs(hostIDs []string) (string, error) {
	if len(hostIDs) == 0 {
		return "", nil
	}

	data, err := json.Marshal(hostIDs)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
