package store

import (
	"VPSBenchmarkBackend/internal/common"
	"VPSBenchmarkBackend/internal/inspector/model"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	qdb "github.com/questdb/go-questdb-client/v4"
	"log"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

var sender qdb.LineSender
var pgConn *pgxpool.Pool
var validIntervalRegex = regexp.MustCompile(`^[1-9]\d*[smhd]$`)

const (
	TrafficMeasurement  = "traffic"
	PingMeasurement     = "ping"
	PointsLimit         = 288
	RetentionDays       = 90
	CleanupScanInterval = 12 * time.Hour
)

func init() {
	var err error
	sender, pgConn, err = buildQuestDB()
	if err != nil {
		panic("failed to initialize QuestDB client: " + err.Error())
	}
	if err = ensureTables(); err != nil {
		panic("failed to initialize QuestDB tables: " + err.Error())
	}
	common.RegisterCronJob(CleanupScanInterval, retentionCleaner)
}

func SaveTrafficPoints(points []model.TrafficPoint) error {
	if len(points) == 0 {
		return nil
	}
	for _, point := range points {
		err := sender.Table(TrafficMeasurement).
			Symbol("host_id", strconv.FormatInt(point.HostID, 10)).
			Float64Column("recv", float64(point.Recv)).
			Float64Column("sent", float64(point.Sent)).
			At(context.Background(), point.Time)
		if err != nil {
			return err
		}
	}
	return sender.Flush(context.Background())
}

func SavePingPoints(points []model.PingPoint) error {
	if len(points) == 0 {
		return nil
	}
	for _, point := range points {
		err := sender.Table(PingMeasurement).
			Symbol("host_id", strconv.FormatInt(point.HostID, 10)).
			Float64Column("latency", float64(point.Latency)).
			At(context.Background(), point.Time)
		if err != nil {
			return err
		}
	}
	return sender.Flush(context.Background())
}

func QueryTrafficSum(hostID int64, start, end int64) (float64, float64, error) {
	row, err := pgConn.Query(context.Background(),
		"SELECT SUM(recv) AS recv_sum, SUM(sent) AS sent_sum FROM "+TrafficMeasurement+" WHERE host_id = $1 AND ts >= $2 AND ts <= $3",
		strconv.FormatInt(hostID, 10),
		time.Unix(0, start).UTC(),
		time.Unix(0, end).UTC(),
	)
	if err != nil {
		return 0, 0, err
	}
	defer row.Close()
	if row.Next() {
		var recvSum, sentSum sql.NullFloat64
		if err = row.Scan(&recvSum, &sentSum); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return 0, 0, nil
			}
			return 0, 0, err
		}
		if !recvSum.Valid || !sentSum.Valid {
			return 0, 0, nil
		}
		return recvSum.Float64, sentSum.Float64, nil
	}
	return 0, 0, row.Err()
}

func QueryLatestPing(hostID int64, start, end int64) (float32, error) {
	row, err := pgConn.Query(context.Background(),
		"SELECT latency FROM "+PingMeasurement+" WHERE host_id = $1 AND ts >= $2 AND ts <= $3 ORDER BY ts DESC LIMIT 1",
		strconv.FormatInt(hostID, 10),
		time.Unix(0, start).UTC(),
		time.Unix(0, end).UTC(),
	)
	if err != nil {
		return 0, err
	}
	defer row.Close()
	var latency sql.NullFloat64
	if row.Next() {
		if err = row.Scan(&latency); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return 0, nil
			}
			return 0, err
		}
	}

	if !latency.Valid {
		return 0, nil
	}
	return float32(latency.Float64), nil
}

func QueryLossRate(hostID int64, start, end int64) (float64, error) {
	row, err := pgConn.Query(context.Background(),
		"SELECT COUNT(*) AS total, SUM(CASE WHEN latency = 0 THEN 1 ELSE 0 END) AS loss FROM "+PingMeasurement+" WHERE host_id = $1 AND ts >= $2 AND ts <= $3",
		strconv.FormatInt(hostID, 10),
		time.Unix(0, start).UTC(),
		time.Unix(0, end).UTC(),
	)
	if err != nil {
		return 0, err
	}
	defer row.Close()
	if !row.Next() {
		return 0, nil
	}
	var total float64
	var loss sql.NullInt64
	if err := row.Scan(&total, &loss); err != nil {
		return 0, err
	}
	if total == 0 || !loss.Valid {
		return 0, nil
	}
	return float64(loss.Int64) / total, nil
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
	if n <= 0 {
		return []model.PingPoint{}, nil
	}
	rows, err := pgConn.Query(context.Background(),
		fmt.Sprintf("SELECT latency, ts FROM %s WHERE host_id = $1 ORDER BY ts ASC LIMIT %d", PingMeasurement, n),
		strconv.FormatInt(hostID, 10),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	points := make([]model.PingPoint, 0, int(n))
	for rows.Next() {
		var latency float64
		var ts time.Time
		if err := rows.Scan(&latency, &ts); err != nil {
			return nil, err
		}
		points = append(points, model.PingPoint{
			HostID:  hostID,
			Latency: float32(latency),
			Time:    ts,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return points, nil
}

func queryLossPoints(hostID int64, start, end int64) ([]model.PingPoint, error) {
	rows, err := pgConn.Query(context.Background(),
		"SELECT ts FROM "+PingMeasurement+" WHERE host_id = $1 AND latency = 0 AND ts >= $2 AND ts <= $3 ORDER BY ts ASC",
		strconv.FormatInt(hostID, 10),
		time.Unix(0, start).UTC(),
		time.Unix(0, end).UTC(),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	points := make([]model.PingPoint, 0)
	for rows.Next() {
		var ts time.Time
		if err := rows.Scan(&ts); err != nil {
			return nil, err
		}
		points = append(points, model.PingPoint{
			HostID:  hostID,
			Latency: 0,
			Time:    ts,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return points, nil
}

func queryLatencyPoints(hostID int64, start, end int64, interval string) ([]model.PingPoint, error) {
	query := fmt.Sprintf(
		"SELECT ts, AVG(latency) AS mean FROM %s WHERE host_id = $1 AND latency > 0 AND ts >= $2 AND ts <= $3 SAMPLE BY %s ALIGN TO CALENDAR LIMIT %d",
		PingMeasurement,
		interval,
		PointsLimit,
	)
	rows, err := pgConn.Query(context.Background(), query, strconv.FormatInt(hostID, 10), time.Unix(0, start).UTC(), time.Unix(0, end).UTC())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	points := make([]model.PingPoint, 0)
	for rows.Next() {
		var ts time.Time
		var mean sql.NullFloat64
		if err := rows.Scan(&ts, &mean); err != nil {
			return nil, err
		}
		if !mean.Valid {
			continue
		}
		points = append(points, model.PingPoint{
			HostID:  hostID,
			Latency: float32(mean.Float64),
			Time:    ts,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return points, nil
}

func ensureTables() error {
	if _, err := pgConn.Exec(context.Background(), fmt.Sprintf(
		"CREATE TABLE IF NOT EXISTS %s (host_id SYMBOL, recv DOUBLE, sent DOUBLE, ts TIMESTAMP) TIMESTAMP(ts) PARTITION BY DAY WAL",
		TrafficMeasurement,
	)); err != nil {
		return err
	}
	if _, err := pgConn.Exec(context.Background(), fmt.Sprintf(
		"CREATE TABLE IF NOT EXISTS %s (host_id SYMBOL, latency DOUBLE, ts TIMESTAMP) TIMESTAMP(ts) PARTITION BY DAY WAL",
		PingMeasurement,
	)); err != nil {
		return err
	}
	return nil
}

func buildQuestDB() (qdb.LineSender, *pgxpool.Pool, error) {
	host := getEnv("QUESTDB_HOST", "127.0.0.1")
	httpPort := getEnv("QUESTDB_HTTP_PORT", "9000")
	pgPort := getEnv("QUESTDB_PG_PORT", "8812")
	user := getEnv("QUESTDB_USER", "admin")
	pass := getEnv("QUESTDB_PASSWORD", "quest")
	log.Printf("Connecting to QuestDB at %s:%s with user %s", host, pgPort, user)
	ctx := context.Background()
	conn, err := pgxpool.New(ctx, fmt.Sprintf("postgresql://%s:%s@%s:%s/qdb", user, pass, host, pgPort))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to QuestDB: %w", err)
	}
	if err = conn.Ping(ctx); err != nil {
		return nil, nil, fmt.Errorf("failed to ping QuestDB: %w", err)
	}
	qdbSender, err := qdb.NewLineSender(context.Background(), qdb.WithHttp(), qdb.WithAddress(fmt.Sprintf("%s:%s", host, httpPort)), qdb.WithBasicAuth(user, pass))
	if err != nil {
		conn.Close()
		return nil, nil, fmt.Errorf("failed to create QuestDB line sender: %w", err)
	}
	return qdbSender, conn, err
}

func getEnv(key string, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func retentionCleaner() {
	// 计算 90 天前的时间
	retentionDate := time.Now().AddDate(0, 0, -RetentionDays).Format("2006-01-02")

	// 执行删除 SQL
	queries := []string{
		fmt.Sprintf("ALTER TABLE %s DROP PARTITION WHERE ts < '%s'", TrafficMeasurement, retentionDate),
		fmt.Sprintf("ALTER TABLE %s DROP PARTITION WHERE ts < '%s'", PingMeasurement, retentionDate),
	}
	_, err := pgConn.Exec(context.Background(), strings.Join(queries, "; "))
	if err != nil {
		log.Printf("Failed to clean expired data: %v", err)
	} else {
		log.Printf("Data before %s was cleaned successfully.", retentionDate)
	}
}
