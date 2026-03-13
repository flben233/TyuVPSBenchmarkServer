package store

import (
	"VPSBenchmarkBackend/internal/inspector/model"
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
	query := fmt.Sprintf("SELECT latency FROM %s WHERE host_id = $host_id ORDER BY time DESC LIMIT %d", PingMeasurement, n)
	params := influxdb3.QueryParameters{
		"host_id": strconv.FormatInt(hostID, 10),
	}
	iter, err := client.QueryWithParameters(ctx, query, params, influxdb3.WithQueryType(influxdb3.InfluxQL))
	if err != nil {
		return nil, err
	}
	var points []model.PingPoint
	for iter.Next() {
		value := iter.Value()
		points = append(points, model.PingPoint{
			HostID:  hostID,
			Latency: float32(value["latency"].(float64)),
			Time:    value["time"].(time.Time),
		})
	}
	if iter.Err() != nil {
		return nil, iter.Err()
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
	iter, err := client.QueryWithParameters(ctx, query, params, influxdb3.WithQueryType(influxdb3.InfluxQL))
	if err != nil {
		return nil, err
	}
	var points []model.PingPoint
	for iter.Next() {
		v := iter.Value()
		points = append(points, model.PingPoint{
			HostID:  hostID,
			Latency: 0,
			Time:    v["time"].(time.Time),
		})
	}
	if iter.Err() != nil {
		return nil, iter.Err()
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
	iter, err := client.QueryWithParameters(ctx, query, params, influxdb3.WithQueryType(influxdb3.InfluxQL))
	if err != nil {
		return nil, err
	}
	var points []model.PingPoint
	for iter.Next() {
		v := iter.Value()
		if v["mean"] != nil {
			points = append(points, model.PingPoint{
				HostID:  hostID,
				Latency: float32(v["mean"].(float64)),
				Time:    time.Unix(0, int64(v["time"].(arrow.Timestamp))), // 从没见过这么丑陋的框架，插入的时候允许time.Time，取出又变成了arrow.Timestamp
			})
		}
	}
	if iter.Err() != nil {
		return nil, iter.Err()
	}
	return points, nil
}
