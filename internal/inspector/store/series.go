package store

import (
	"VPSBenchmarkBackend/internal/inspector/model"
	"context"
	"fmt"
	"github.com/InfluxCommunity/influxdb3-go/v2/influxdb3"
)

var client *influxdb3.Client
var ctx = context.Background()

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
	pArr := make([]any, len(points))
	for i, point := range points {
		p := struct {
			Measurement string `influx:"measurement"`
			model.TrafficPoint
		}{
			Measurement:  TrafficMeasurement,
			TrafficPoint: point,
		}
		pArr[i] = p
	}
	return client.WriteData(ctx, pArr)
}

func SavePingPoints(points []model.PingPoint) error {
	pArr := make([]any, len(points))
	for i, point := range points {
		p := struct {
			Measurement string `influx:"measurement"`
			model.PingPoint
		}{
			Measurement: PingMeasurement,
			PingPoint:   point,
		}
		pArr[i] = p
	}
	return client.WriteData(ctx, pArr)
}

func QueryTrafficPoints(hostID int64, start, end int64, interval string) ([]model.TrafficPoint, error) {
	return queryPoints(TrafficMeasurement, hostID, start, end, interval, func(v map[string]interface{}) model.TrafficPoint {
		return model.TrafficPoint{
			HostID:      hostID,
			Consumption: v["mean"].(float32),
			Time:        v["time"].(int64),
		}
	})
}

func QueryPingPoints(hostID int64, start, end int64, interval string) ([]model.PingPoint, error) {
	return queryPoints(PingMeasurement, hostID, start, end, interval, func(v map[string]interface{}) model.PingPoint {
		return model.PingPoint{
			HostID:  hostID,
			Latency: v["mean"].(float32),
			Time:    v["time"].(int64),
		}
	})
}

func queryPoints[T any](measurement string, hostID int64, start, end int64, interval string, handler func(v map[string]interface{}) T) ([]T, error) {
	query := fmt.Sprintf("SELECT MEAN(*) FROM %s WHERE host_id = $host_id AND time >= $start AND time <= $end GROUP BY time($interval) LIMIT 120", measurement)
	params := influxdb3.QueryParameters{
		"host_id":  hostID,
		"start":    start,
		"end":      end,
		"interval": interval,
	}
	iter, err := client.QueryWithParameters(ctx, query, params, influxdb3.WithQueryType(influxdb3.InfluxQL))
	if err != nil {
		return nil, err
	}
	var points []T
	for iter.Next() {
		value := iter.Value()
		points = append(points, handler(value))
	}
	if iter.Err() != nil {
		return nil, iter.Err()
	}
	return points, nil
}
