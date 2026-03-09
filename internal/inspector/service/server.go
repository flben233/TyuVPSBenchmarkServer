package service

import (
	"VPSBenchmarkBackend/internal/auth/util"
	"VPSBenchmarkBackend/internal/cache"
	"VPSBenchmarkBackend/internal/common"
	"VPSBenchmarkBackend/internal/config"
	"VPSBenchmarkBackend/internal/inspector/model"
	"VPSBenchmarkBackend/internal/inspector/response"
	"VPSBenchmarkBackend/internal/inspector/store"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
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
	pingInterval   = 120 * time.Second
	putInterval    = 30 * time.Second
	putMaxLength   = 1
	notifyInterval = 5 * time.Minute
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

		if lat == 0 && host.Notify {
			tryNotify(host.UserID, &host, fmt.Sprintf("主机 %s (%s) 离线", host.Name, host.Target))
		} else if lat > 0 {
			ping, err := store.QueryLatestPing(host.ID)
			if err != nil {
				log.Printf("Failed to query latest ping for host %d: %v", host.ID, err)
			} else if ping == 0 && host.Notify {
				tryNotify(host.UserID, &host, fmt.Sprintf("主机 %s (%s) 上线", host.Name, host.Target))
			}
		}

		if err := store.SavePingPoints(pingData); err != nil {
			log.Printf("Failed to save ping points for host %d: %v", host.ID, err)
		}
	}
}

func CreateHost(userID int64, target, name, tags string, notify bool) (int64, error) {
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
		ID:     id,
		UserID: userID,
		Target: target,
		Name:   name,
		Tags:   tags,
		Notify: notify,
	}
	if err := store.CreateHost(host); err != nil {
		return 0, err
	}
	return host.ID, nil
}

func UpdateHost(userID int64, hostID int64, name, tags, target string, notify bool) error {
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
			ServerStatus: host.ServerStatus,
		}
	}
	return inspectHosts, nil
}

func PutData(trafficData []model.TrafficPoint, hostInfo common.ServerStatus, hostID int64) error {
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

	host = &model.InspectHost{
		ID:           host.ID,
		UserID:       host.UserID,
		Target:       host.Target,
		Name:         host.Name,
		Tags:         host.Tags,
		ServerStatus: hostInfo,
	}

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
		recv, sent, err := store.QueryTrafficSum(host.ID, start, end, interval)
		if err != nil {
			return nil, err
		}
		latestPing, err := store.QueryLatestPing(host.ID)
		if err != nil {
			return nil, err
		}
		data[i] = &response.HostData{
			Ping:         pingPoints,
			Sent:         sent,
			Recv:         recv,
			ID:           strconv.FormatInt(host.ID, 10),
			Target:       host.Target,
			Name:         host.Name,
			Tags:         host.Tags,
			Notify:       host.Notify,
			LatestPing:   latestPing,
			ServerStatus: host.ServerStatus,
		}
	}
	return data, nil
}

func GetUserSettings(userID int64) *response.SettingData {
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
	return &response.SettingData{
		NotifyURL: setting.NotifyURL,
		BgURL:     setting.BgURL,
	}
}

func UpdateUserSettings(userID int64, notifyURL, bgURL *string) error {
	setting, err := store.GetSettingByUserID(userID)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		setting = &model.InspectorSetting{UserID: userID}
	}
	setting.BgURL = bgURL
	setting.NotifyURL = notifyURL
	return store.UpsertSetting(setting)
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

func tryNotify(userID int64, host *model.InspectHost, message string) {
	setting, err := store.GetSettingByUserID(userID)
	if err != nil {
		log.Printf("Failed to get setting for user %d: %v", userID, err)
	}
	if setting.NotifyURL == nil {
		return
	}
	hostID := host.ID
	err = cache.GetClient().Get(context.Background(), strconv.FormatInt(hostID, 10)).Err()
	if errors.Is(err, redis.Nil) {
		// 发送通知
		go func() {
			notifyReq := map[string]interface{}{
				"urls":  setting.NotifyURL,
				"body":  message,
				"title": "Lolicon Monitor 通知",
			}
			reqBytes, err := json.Marshal(notifyReq)
			if err != nil {
				log.Printf("Failed to marshal notify request for host %d: %v", hostID, err)
				return
			}
			resp, err := http.Post(config.Get().AppriseURL, "application/json", bytes.NewReader(reqBytes))
			if err != nil {
				log.Printf("Failed to send notify request for host %d: %v", hostID, err)
				return
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				bodyBytes, _ := io.ReadAll(resp.Body)
				log.Printf("Failed to send notify request for host %d, status: %d, response: %s", hostID, resp.StatusCode, string(bodyBytes))
			}
		}()
		// 设置Redis键，过期时间为通知间隔，防止频繁通知
		cache.GetClient().Set(context.Background(), strconv.FormatInt(hostID, 10), setting.NotifyURL, notifyInterval)
	}
}
