package store

import (
	"VPSBenchmarkBackend/internal/inspector/model"
	"context"
	"fmt"
	"github.com/InfluxCommunity/influxdb3-go/v2/influxdb3"
	"github.com/apache/arrow-go/v18/arrow"
	"regexp"
	"strconv"
	"time"
)

var client *influxdb3.Client
var ctx = context.Background()
var validIntervalRegex = regexp.MustCompile(`^[1-9]\d*[smhd]$`)

const (
	TrafficMeasurement = "traffic"
	PingMeasurement    = "ping"
)

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

func QueryTrafficSum(hostID int64, start, end int64, interval string) (float64, float64, error) {
	match := validIntervalRegex.Match([]byte(interval))
	if !match {
		return 0, 0, fmt.Errorf("invalid interval format: %s", interval)
	}
	query := fmt.Sprintf("SELECT SUM(recv) AS recv_sum, SUM(sent) AS sent_sum FROM %s WHERE host_id = $host_id AND time >= $start AND time <= $end GROUP BY time(%s)", TrafficMeasurement, interval)
	params := influxdb3.QueryParameters{
		"host_id": strconv.FormatInt(hostID, 10),
		"start":   time.Unix(0, start),
		"end":     time.Unix(0, end),
	}
	iter, err := client.QueryWithParameters(ctx, query, params, influxdb3.WithQueryType(influxdb3.InfluxQL))
	if err != nil {
		return 0, 0, err
	}
	var recvSum, sentSum float64
	for iter.Next() {
		value := iter.Value()
		if value["recv_sum"] == nil || value["sent_sum"] == nil {
			continue
		}
		recvSum += value["recv_sum"].(float64)
		sentSum += value["sent_sum"].(float64)
	}
	if iter.Err() != nil {
		return 0, 0, iter.Err()
	}
	return recvSum, sentSum, nil
}

func QueryLatestPing(hostID int64) (float32, error) {
	query := fmt.Sprintf("SELECT latency FROM %s WHERE host_id = $host_id ORDER BY time DESC LIMIT 1", PingMeasurement)
	params := influxdb3.QueryParameters{
		"host_id": strconv.FormatInt(hostID, 10),
	}
	iter, err := client.QueryWithParameters(ctx, query, params, influxdb3.WithQueryType(influxdb3.InfluxQL))
	if err != nil {
		return 0, err
	}
	var latency float32
	if iter.Next() {
		value := iter.Value()
		latency = float32(value["latency"].(float64))
	}
	if iter.Err() != nil {
		return 0, iter.Err()
	}
	return latency, nil
}

func QueryPingPoints(hostID int64, start, end int64, interval string) ([]model.PingPoint, error) {
	return queryPoints(PingMeasurement, hostID, start, end, interval, func(v map[string]interface{}) model.PingPoint {
		return model.PingPoint{
			HostID:  hostID,
			Latency: float32(v["mean"].(float64)),
			Time:    time.Unix(0, int64(v["time"].(arrow.Timestamp))), // 从没见过这么丑陋的框架，插入的时候允许time.Time，取出又变成了arrow.Timestamp
		}
	})
}

func queryPoints[T any](measurement string, hostID int64, start, end int64, interval string, handler func(v map[string]interface{}) T) ([]T, error) {
	match := validIntervalRegex.Match([]byte(interval))
	if !match {
		return nil, fmt.Errorf("invalid interval format: %s", interval)
	}
	query := fmt.Sprintf("SELECT MEAN(latency) FROM %s WHERE host_id = $host_id AND time >= $start AND time <= $end GROUP BY time(%s) LIMIT 120", measurement, interval)
	params := influxdb3.QueryParameters{
		"host_id": strconv.FormatInt(hostID, 10),
		"start":   time.Unix(0, start),
		"end":     time.Unix(0, end),
	}
	iter, err := client.QueryWithParameters(ctx, query, params, influxdb3.WithQueryType(influxdb3.InfluxQL))
	if err != nil {
		return nil, err
	}
	var points []T
	for iter.Next() {
		value := iter.Value()
		if value["mean"] != nil {
			points = append(points, handler(value))
		}
	}
	if iter.Err() != nil {
		return nil, iter.Err()
	}
	return points, nil
}
