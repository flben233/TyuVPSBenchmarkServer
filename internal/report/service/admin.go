package service

import (
	"VPSBenchmarkBackend/internal/report/model"
	"VPSBenchmarkBackend/internal/report/parser"
	"VPSBenchmarkBackend/internal/report/store"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// generateID generates a random ID for reports
func generateID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// AddReport parses and saves a new benchmark report
func AddReport(rawHTML string) (string, error) {
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

	convertMediaIndex(false, parsedResult.Media.IPv4)
	convertMediaIndex(true, parsedResult.Media.IPv6)
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
	for key := range parsedResult.ECS.Info {
		if strings.Contains(key, "IPV6") {
			ipv6Support = true
		}
		if strings.Contains(key, "虚拟化") {
			virtualization = parsedResult.ECS.Info[key]
		}
	}
	seqRead, err := strconv.ParseFloat(parsedResult.Disk.Data[0][1], 32)
	if err != nil {
		return "", fmt.Errorf("failed to parse disk sequential read speed: %w", err)
	}
	seqWrite, err := strconv.ParseFloat(parsedResult.Disk.Data[0][2], 32)
	if err != nil {
		return "", fmt.Errorf("failed to parse disk sequential write speed: %w", err)
	}
	ei = model.InfoIndex{
		ReportID:       reportID,
		IPv6Support:    ipv6Support,
		Virtualization: virtualization,
		SeqRead:        float32(seqRead),
		SeqWrite:       float32(seqWrite),
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
		CreatedAt: currentTime,
		UpdatedAt: currentTime,
	}

	// Save to database
	if err := store.SaveReport(report, mi, si, &ei, bi); err != nil {
		return "", fmt.Errorf("failed to save report: %w", err)
	}

	return reportID, nil
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
