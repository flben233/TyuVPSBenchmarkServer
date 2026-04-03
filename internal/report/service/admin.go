package service

import (
	"VPSBenchmarkBackend/internal/mq"
	"VPSBenchmarkBackend/internal/report/model"
	"VPSBenchmarkBackend/internal/report/parser"
	"VPSBenchmarkBackend/internal/report/request"
	"VPSBenchmarkBackend/internal/report/store"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
	"strconv"
	"strings"
	"time"
)

const reportRoute = "report_processing"

func init() {
	// Subscribe to mq for report processing
	mq.LateSubscribe(reportRoute, context.Background(), processReports)
}

// generateID generates a random ID for reports
func generateID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func processReports(msg *amqp.Delivery) error {
	var reqArr []request.AddReportRequest
	err := json.Unmarshal(msg.Body, &reqArr)
	if err != nil {
		return fmt.Errorf("failed to unmarshal report processing request: %w", err)
	}

	task, err := mq.GetTask[model.AddReportTask](msg.MessageId)
	if err != nil {
		return fmt.Errorf("failed to get report task from Redis: %w", err)
	}

	task.Status = mq.TaskRunning
	err = mq.SetTask(*task)
	if err != nil {
		return fmt.Errorf("failed to update report task status in Redis: %w", err)
	}

	for i, req := range reqArr {
		_, err = addReport(req.HTML, req.MonitorID, req.OtherInfo)

		if err != nil {
			log.Println("Error processing report:", err)
			task.Result.Failed = append(task.Result.Failed, i)
		}
		task.Progress = float32(i+1) / float32(len(reqArr))
		if i%2 == 0 { // Update progress every 2 reports to reduce Redis writes
			err = mq.SetTask(*task)
			if err != nil {
				log.Printf("Failed to update report task progress in Redis: %v", err)
			}
		}
	}
	task.Status = mq.TaskDone
	task.Progress = 1.0

	return mq.SetTask(*task)
}

// addReport parses and saves a new benchmark report
func addReport(rawHTML string, monitorID *int64, otherInfo string) (string, error) {
	if rawHTML == "" {
		return "", errors.New("raw HTML content is required")
	}
	// Generate unique ID
	reportID := generateID()

	// Parse the report
	parsedResult := parser.MainParser(rawHTML)
	mi := make([]model.MediaIndex, 0)
	si := make([]model.SpeedtestIndex, 0)
	bi := make([]model.BacktraceIndex, 0)
	var ei model.InfoIndex

	var convertMediaIndex = func(isIPv6 bool, mediaBlocks []model.MediaBlock) {
		for _, media := range mediaBlocks {
			region := media.Region
			for _, m := range media.Results {
				mi = append(mi, model.MediaIndex{
					ReportID: reportID,
					Region:   region,
					Media:    m.Media,
					Unlock:   strings.Contains(m.Unlock, "Yes"),
					IPv6:     isIPv6,
				})
			}
		}
	}

	if parsedResult.Media != nil {
		convertMediaIndex(false, parsedResult.Media.IPv4)
		convertMediaIndex(true, parsedResult.Media.IPv6)
	}

	for _, st := range parsedResult.SpdTest {
		for _, r := range st.Results {
			isp := ""
			if strings.Contains(r.Spot, "电信") {
				isp = model.ISPChinaTelecom
			} else if strings.Contains(r.Spot, "联通") {
				isp = model.ISPChinaUnicom
			} else if strings.Contains(r.Spot, "移动") {
				isp = model.ISPChinaMobile
			}
			si = append(si, model.SpeedtestIndex{
				ReportID: reportID,
				Spot:     r.Spot,
				Download: r.Download,
				Upload:   r.Upload,
				Latency:  r.Latency,
				ISP:      isp,
			})
		}
	}

	ipv6Support := false
	virtualization := ""
	if parsedResult.ECS != nil {
		// Parse IPv6 support and virtualization from ECS.Info
		for key := range parsedResult.ECS.Info {
			if strings.Contains(key, "IPV6") {
				ipv6Support = true
			}
			if strings.Contains(key, "虚拟化") {
				virtualization = parsedResult.ECS.Info[key]
			}
		}
		// Parse backtrace data from ECS.Trace.Types
		// Filter items where value contains "线路" (route type)
		for spot, routeType := range parsedResult.ECS.Trace.Types {
			if strings.Contains(routeType, "线路") {
				isp := ""
				if strings.Contains(spot, "电信") {
					isp = model.ISPChinaTelecom
				} else if strings.Contains(spot, "联通") {
					isp = model.ISPChinaUnicom
				} else if strings.Contains(spot, "移动") {
					isp = model.ISPChinaMobile
				}
				bi = append(bi, model.BacktraceIndex{
					ReportID:  reportID,
					Spot:      spot,
					RouteType: routeType,
					ISP:       isp,
				})
			}
		}
		// Parse media unlock from ECS.Tiktok
		if strings.Contains(parsedResult.ECS.Tiktok, "【") {
			i1 := model.MediaIndex{
				ReportID: reportID,
				Region:   strings.Trim(parsedResult.ECS.Tiktok, "【】"),
				Media:    "TikTok",
				Unlock:   true,
				IPv6:     false,
			}
			mi = append(mi, i1)
			i1.IPv6 = true
			mi = append(mi, i1)
		}
	}

	var seqRead, seqWrite float64
	var err error
	if parsedResult.Disk != nil && len(parsedResult.Disk.Data) > 0 && len(parsedResult.Disk.Data[0]) == 3 {
		seqRead, err = strconv.ParseFloat(parsedResult.Disk.Data[0][1], 32)
		if err != nil {
			return "", fmt.Errorf("failed to parse disk sequential read speed: %w", err)
		}
		seqWrite, err = strconv.ParseFloat(parsedResult.Disk.Data[0][2], 32)
		if err != nil {
			return "", fmt.Errorf("failed to parse disk sequential write speed: %w", err)
		}
	}
	ei = model.InfoIndex{
		ReportID:       reportID,
		IPv6Support:    ipv6Support,
		Virtualization: virtualization,
		SeqRead:        float32(seqRead),
		SeqWrite:       float32(seqWrite),
	}

	// Check if report already exists
	exists, err := store.ReportExists(reportID)
	if err != nil {
		return "", fmt.Errorf("failed to check report existence: %w", err)
	}
	if exists {
		// Retry with a new ID
		reportID = generateID()
	}

	currentTime := time.Now()
	// Create BenchmarkResult for database
	report := &model.BenchmarkResult{
		ReportID:  reportID,
		Title:     parsedResult.Title,
		Time:      parsedResult.Time,
		Link:      parsedResult.Link,
		SpdTest:   parsedResult.SpdTest,
		ECS:       parsedResult.ECS,
		Media:     parsedResult.Media,
		BestTrace: parsedResult.BestTrace,
		Itdog:     parsedResult.Itdog,
		Disk:      parsedResult.Disk,
		IPQuality: parsedResult.IPQuality,
		MonitorID: monitorID,
		OtherInfo: otherInfo,
		CreatedAt: currentTime,
		UpdatedAt: currentTime,
	}

	// Save to database
	if err := store.SaveReport(report, mi, si, &ei, bi); err != nil {
		return "", fmt.Errorf("failed to save report: %w", err)
	}

	return reportID, nil
}

func AddReportsAsync(request []request.AddReportRequest) (string, error) {
	id, err := mq.PublishJSONWithID(reportRoute, "", request, "")
	if err != nil {
		return "", fmt.Errorf("failed to enqueue report processing task: %w", err)
	}
	task := mq.Task[model.AddReportTask]{
		ID:       id,
		Status:   mq.TaskPending,
		Progress: 0,
		Result:   model.AddReportTask{Failed: make([]int, 0)},
	}
	err = mq.SetTask(task)
	if err != nil {
		return "", fmt.Errorf("failed to marshal report task data: %w", err)
	}
	return id, nil
}

// DeleteReport removes a report from the database
func DeleteReport(reportID string) error {
	if reportID == "" {
		return errors.New("report ID is required")
	}

	// Check if report exists
	exists, err := store.ReportExists(reportID)
	if err != nil {
		return fmt.Errorf("failed to check report existence: %w", err)
	}
	if !exists {
		return errors.New("report not found")
	}

	// Delete the report and all related data
	if err := store.DeleteReport(reportID); err != nil {
		return fmt.Errorf("failed to delete report: %w", err)
	}

	return nil
}

func UpdateReport(reportID string, monitorID *int64, otherInfo string) error {
	if reportID == "" {
		return errors.New("report ID is required")
	}

	err := store.UpdateReport(reportID, monitorID, otherInfo)
	if err != nil {
		return fmt.Errorf("failed to update report monitor ID: %w", err)
	}
	return nil
}
