package frontreport

// BatchReportStorage is a way to store incoming reports
type BatchReportStorage interface {
	AddReport(Reportable)
}

// Logger is a simple logging wrapper interface
type Logger interface {
	Log(...interface{}) error
}

// MetricStorage is a way to store internal application metrics
type MetricStorage interface {
	RegisterHistogram(string)
	UpdateHistogram(string, int)
	RegisterCounter(string)
	IncCounter(string, int)
}

// Service is started and stopped in main function, which assembles services into a working application
type Service interface {
	Start() error
	Stop() error
}
