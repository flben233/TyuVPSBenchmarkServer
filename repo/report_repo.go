package repo

import (
	"VPSBenchmarkBackend/model"
	"database/sql"
	"fmt"
)

func ParseReportInfo(rows *sql.Rows) ([]model.ReportInfo, error) {
	var infos []model.ReportInfo
	for rows.Next() {
		var title, path, date string
		if err := rows.Scan(&title, &path, &date); err != nil {
			return nil, err
		}
		infos = append(infos, model.ReportInfo{title, path, date})
	}
	return infos, nil
}

func FindReportsByConditions(keyword string, upload float32, rtype string) ([]model.ReportInfo, error) {
	db := GetDatabase()
	defer db.Close()

	rows, err := db.Query(`
		select distinct id, path, date 
		from report, speedtest, route 
		where report.id = speedtest.rid 
			and report.id = route.rid 
			and report.id like ? 
			and upload >= ?
			and rtype like ?`, "%"+keyword+"%", upload, "%"+rtype+"%")
	if err != nil {
		return nil, fmt.Errorf("failed to query reports: %w", err)
	}
	return ParseReportInfo(rows)
}

func CascadeDeleteReport(title string) error {
	db := GetDatabase()
	defer db.Close()
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
	}

	if _, err := tx.Exec("delete from speedtest where rid = ?", title); err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete speedtest: %w", err)
	}
	if _, err := tx.Exec("delete from route where rid = ?", title); err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete route: %w", err)
	}
	if _, err := tx.Exec("delete from report where id = ?", title); err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete report: %w", err)
	}
	if err := tx.Commit(); err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil
}

func CascadeInsertReport(report model.ReportInfo, speedtests []model.SpeedtestResult, routes model.TraceResult) error {
	db := GetDatabase()
	defer db.Close()
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
	}

	if _, err := tx.Exec("insert or replace into report (id, path, date) values (?, ?, ?)", report.Name, report.Path, report.Date); err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to insert or replace report: %w", err)
	}

	for _, speedtest := range speedtests {
		if _, err := tx.Exec("insert or replace into speedtest (spot, download, upload, rid) values (?, ?, ?, ?)", speedtest.Spot, speedtest.Download, speedtest.Upload, report.Name); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to insert or replace speedtest: %w", err)
		}
	}

	for spot, rtype := range routes.Types {
		if _, err := tx.Exec("insert or replace into route (spot, rtype, rid) values (?, ?, ?)", spot, rtype, report.Name); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to insert or replace route: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil
}
