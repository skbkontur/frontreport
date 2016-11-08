package frontreport

// BatchReportStorage is a way to store incoming reports
type BatchReportStorage interface {
	AddReport(Reportable)
}

// Logger is a simple logging wrapper interface
type Logger interface {
	Log(...interface{}) error
}

// Service is started and stopped in main function, which assembles services into a working application
type Service interface {
	Start() error
	Stop() error
}
