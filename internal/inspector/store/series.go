package store

import (
	"VPSBenchmarkBackend/internal/inspector/model"
	"VPSBenchmarkBackend/pkg/influxdb"
	"context"
	"fmt"
	"github.com/InfluxCommunity/influxdb3-go/v2/influxdb3"
	"github.com/apache/arrow-go/v18/arrow"
	"regexp"
	"sort"
	"strconv"
	"time"
)

var client *influxdb3.Client
var ctx = context.Background()
var validIntervalRegex = regexp.MustCompile(`^[1-9]\d*[smhd]$`)

const (
	TrafficMeasurement = "traffic"
	PingMeasurement    = "ping"
	PointsLimit        = 288
)

type traffic struct {
	RecvSum float64 `json:"recv_sum"`
	SentSum float64 `json:"sent_sum"`
}

func init() {
	// Initialize the InfluxDB client
	client1, err := influxdb3.NewFromEnv()
	if err != nil {
		panic("Failed to initialize InfluxDB client: " + err.Error())
	}
	client = client1
}

func SaveTrafficPoints(points []model.TrafficPoint) error {
	// Create a new point with the appropriate measurement and tags
	pArr := make([]*influxdb3.Point, len(points))
	for i, point := range points {
		pArr[i] = influxdb3.NewPoint(
			TrafficMeasurement,
			map[string]string{"host_id": strconv.FormatInt(point.HostID, 10)},
			map[string]interface{}{
				"recv": point.Recv,
				"sent": point.Sent,
			},
			point.Time,
		)
	}
	return client.WritePoints(ctx, pArr)
}

func SavePingPoints(points []model.PingPoint) error {
	pArr := make([]*influxdb3.Point, len(points))
	for i, point := range points {
		pArr[i] = influxdb3.NewPoint(
			PingMeasurement,
			map[string]string{"host_id": strconv.FormatInt(point.HostID, 10)},
			map[string]interface{}{"latency": point.Latency},
			point.Time,
		)
	}
	return client.WritePoints(ctx, pArr)
}

func QueryTrafficSum(hostID int64, start, end int64) (float64, float64, error) {
	query := fmt.Sprintf("SELECT SUM(recv) AS recv_sum, SUM(sent) AS sent_sum FROM %s WHERE host_id = $host_id AND time >= $start AND time <= $end", TrafficMeasurement)
	params := influxdb3.QueryParameters{
		"host_id": strconv.FormatInt(hostID, 10),
		"start":   time.Unix(0, start),
		"end":     time.Unix(0, end),
	}
	t, err := influxdb.QueryItem(query, params, client, func(value map[string]interface{}) (*traffic, error) {
		if value["recv_sum"] == nil || value["sent_sum"] == nil {
			return nil, nil
		}
		return &traffic{
			RecvSum: value["recv_sum"].(float64),
			SentSum: value["sent_sum"].(float64),
		}, nil
	})
	if err != nil || t == nil {
		return 0, 0, err
	}
	return t.RecvSum, t.SentSum, nil
}

func QueryLatestPing(hostID int64, start, end int64) (float32, error) {
	query := fmt.Sprintf("SELECT latency FROM %s WHERE host_id = $host_id AND time >= $start AND time <= $end ORDER BY time DESC LIMIT 1", PingMeasurement)
	params := influxdb3.QueryParameters{
		"host_id": strconv.FormatInt(hostID, 10),
		"start":   time.Unix(0, start),
		"end":     time.Unix(0, end),
	}
	latency, err := influxdb.QueryItem(query, params, client, func(value map[string]interface{}) (*float32, error) {
		f := float32(value["latency"].(float64))
		return &f, nil
	})
	if err != nil {
		return 0, err
	}
	return *latency, nil
}

func QueryLossRate(hostID int64, start, end int64) (float64, error) {
	query := fmt.Sprintf("SELECT COUNT(*) AS total, SUM(CASE WHEN latency = 0 THEN 1 ELSE 0 END) AS loss FROM %s WHERE host_id = $host_id AND time >= $start AND time <= $end", PingMeasurement)
	params := influxdb3.QueryParameters{
		"host_id": strconv.FormatInt(hostID, 10),
		"start":   time.Unix(0, start),
		"end":     time.Unix(0, end),
	}
	rate, err := influxdb.QueryItemWithQL(query, params, client, influxdb3.SQL, func(value map[string]interface{}) (*float64, error) {
		if value["total"] == nil || value["loss"] == nil {
			return nil, nil
		}
		total := value["total"].(int64)
		loss := value["loss"].(int64)
		if total == 0 {
			return new(float64), nil
		}
		rate := float64(loss) / float64(total)
		return &rate, nil
	})
	if err != nil {
		return 0, err
	}
	return *rate, nil
}

func QueryPingPoints(hostID int64, start, end int64, interval string) ([]model.PingPoint, error) {
	match := validIntervalRegex.Match([]byte(interval))
	if !match {
		return nil, fmt.Errorf("invalid interval format: %s", interval)
	}
	latencyPoints, err := queryLatencyPoints(hostID, start, end, interval)
	if err != nil {
		return nil, err
	}
	lossPoints, err := queryLossPoints(hostID, start, end)
	if err != nil {
		return nil, err
	}
	points := append(latencyPoints, lossPoints...)
	sort.Slice(points, func(i, j int) bool {
		return points[i].Time.Before(points[j].Time)
	})
	return points, nil
}

func QueryLatestNPingPoints(hostID int64, n int64) ([]model.PingPoint, error) {
	query := fmt.Sprintf("SELECT latency FROM %s WHERE host_id = $host_id ORDER BY time ASC LIMIT %d", PingMeasurement, n)
	params := influxdb3.QueryParameters{
		"host_id": strconv.FormatInt(hostID, 10),
	}
	points, err := influxdb.QueryItems(query, params, client, func(value map[string]interface{}) (*model.PingPoint, error) {
		return &model.PingPoint{
			HostID:  hostID,
			Latency: float32(value["latency"].(float64)),
			Time:    value["time"].(time.Time),
		}, nil
	})
	if err != nil {
		return nil, err
	}
	return points, nil
}

func queryLossPoints(hostID int64, start, end int64) ([]model.PingPoint, error) {
	query := fmt.Sprintf("SELECT * FROM %s WHERE host_id = $host_id AND latency = 0 AND time >= $start AND time <= $end", PingMeasurement)
	params := influxdb3.QueryParameters{
		"host_id": strconv.FormatInt(hostID, 10),
		"start":   time.Unix(0, start),
		"end":     time.Unix(0, end),
	}
	points, err := influxdb.QueryItems(query, params, client, func(value map[string]interface{}) (*model.PingPoint, error) {
		return &model.PingPoint{
			HostID:  hostID,
			Latency: 0,
			Time:    value["time"].(time.Time),
		}, nil
	})
	if err != nil {
		return nil, err
	}
	return points, nil
}

func queryLatencyPoints(hostID int64, start, end int64, interval string) ([]model.PingPoint, error) {
	query := fmt.Sprintf("SELECT MEAN(latency) FROM %s WHERE host_id = $host_id AND latency > 0 AND time >= $start AND time <= $end GROUP BY time(%s) LIMIT %d", PingMeasurement, interval, PointsLimit)
	params := influxdb3.QueryParameters{
		"host_id": strconv.FormatInt(hostID, 10),
		"start":   time.Unix(0, start),
		"end":     time.Unix(0, end),
	}
	points, err := influxdb.QueryItems(query, params, client, func(value map[string]interface{}) (*model.PingPoint, error) {
		if value["mean"] == nil {
			return nil, nil
		}
		return &model.PingPoint{
			HostID:  hostID,
			Latency: float32(value["mean"].(float64)),
			Time:    time.Unix(0, int64(value["time"].(arrow.Timestamp))),
		}, nil
	})
	if err != nil {
		return nil, err
	}
	return points, nil
}
