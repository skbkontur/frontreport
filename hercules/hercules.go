package hercules

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/skbkontur/frontreport"
)

// ReportStorage is a Hercules implementation of frontreport.ReportStorage interface
type ReportStorage struct {
	Logger           frontreport.Logger
	MetricStorage    frontreport.MetricStorage
	HerculesEndpoint string
	HerculesAPIKey   string
	metrics          struct {
		reportEncodingErrors frontreport.MetricCounter
		adapterRequestTotal  frontreport.MetricCounter
		adapterRequestErrors frontreport.MetricCounter
	}
}

// Start initializes metrics
func (rs *ReportStorage) Start() error {
	rs.metrics.reportEncodingErrors = rs.MetricStorage.RegisterCounter("hercules.report_encoding.errors")
	rs.metrics.adapterRequestTotal = rs.MetricStorage.RegisterCounter("hercules.adapter_request.total")
	rs.metrics.adapterRequestErrors = rs.MetricStorage.RegisterCounter("hercules.adapter_request.errors")

	return nil
}

// Stop does nothing
func (rs *ReportStorage) Stop() error {
	return nil
}

// AddReport directly sends a report to Hercules without any batching
func (rs *ReportStorage) AddReport(report frontreport.Reportable) {
	reportJSON, err := json.Marshal(report)
	if err != nil {
		rs.Logger.Log("msg", "failed to encode", "report_type", report.GetType(), "error", err)
		rs.metrics.reportEncodingErrors.Inc(1)
		return
	}

	var indexName string
	if report.GetService() != "" {
		indexName = fmt.Sprintf("%s-report-%s-%s", report.GetType(), report.GetService(), time.Now().UTC().Format("2006.01.02"))
	} else {
		indexName = fmt.Sprintf("%s-report-%s", report.GetType(), time.Now().UTC().Format("2006.01.02"))
	}

	rs.metrics.adapterRequestTotal.Inc(1)

	client := http.Client{}
	request, err := http.NewRequest(
		"POST",
		fmt.Sprintf("%s/logs/%s", rs.HerculesEndpoint, indexName),
		bytes.NewReader(reportJSON))
	if err != nil {
		rs.Logger.Log(
			"msg", "failed to initialize request to Hercules API",
			"report_type", report.GetType(),
			"error", err)
		rs.metrics.adapterRequestErrors.Inc(1)
		return
	}

	request.Header.Add("Content-Type", "application/json")
	request.Header.Add("Authorization", "ELK "+rs.HerculesAPIKey)

	response, err := client.Do(request)
	if err != nil {
		rs.Logger.Log(
			"msg", "failed to send request to Hercules API",
			"report_type", report.GetType(),
			"error", err)
		rs.metrics.adapterRequestErrors.Inc(1)
		return
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		rs.Logger.Log(
			"msg", "non-200 response code from Hercules API",
			"report_type", report.GetType(),
			"response_code", response.StatusCode)
		rs.metrics.adapterRequestErrors.Inc(1)
	}
}
