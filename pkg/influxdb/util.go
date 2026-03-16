package influxdb

import (
	"context"
	"github.com/InfluxCommunity/influxdb3-go/v2/influxdb3"
)

// QueryItem executes a query and returns the first result item, or nil if no items are found.
// The scanFunc is used to convert the raw query result into the desired type T.
// The params of scanFunc is the raw query result represented as a map[string]interface{}. The scanFunc should return nil if the item should be skipped.
func QueryItem[T any](query string, params influxdb3.QueryParameters, client *influxdb3.Client, scanFunc func(map[string]interface{}) (*T, error)) (*T, error) {
	return QueryItemWithQL(query, params, client, influxdb3.InfluxQL, scanFunc)
}

func QueryItems[T any](query string, params influxdb3.QueryParameters, client *influxdb3.Client, scanFunc func(map[string]interface{}) (*T, error)) ([]T, error) {
	return QueryItemsWithQL(query, params, client, influxdb3.InfluxQL, scanFunc)
}

func QueryItemWithQL[T any](query string, params influxdb3.QueryParameters, client *influxdb3.Client, queryType influxdb3.QueryType, scanFunc func(map[string]interface{}) (*T, error)) (*T, error) {
	items, err := QueryItemsWithQL(query, params, client, queryType, scanFunc)
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, nil
	}
	return &items[0], nil
}

func QueryItemsWithQL[T any](query string, params influxdb3.QueryParameters, client *influxdb3.Client, queryType influxdb3.QueryType, scanFunc func(map[string]interface{}) (*T, error)) ([]T, error) {
	iter, err := client.QueryWithParameters(context.Background(), query, params, influxdb3.WithQueryType(queryType))
	if err != nil {
		return nil, err
	}
	var results []T
	for iter.Next() {
		value := iter.Value()
		item, err := scanFunc(value)
		if err != nil {
			return nil, err
		}
		if item != nil {
			results = append(results, *item)
		}
	}
	if iter.Err() != nil {
		return nil, iter.Err()
	}
	return results, nil
}
