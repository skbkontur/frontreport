package frontreport

// Logger is a simple logging wrapper interface
type Logger interface {
	Log(...interface{}) error
}

// MetricStorage is a way to store internal application metrics
type MetricStorage interface {
	RegisterHistogram(string) MetricHistogram
	RegisterCounter(string) MetricCounter
}

// MetricHistogram is a simple histogram
type MetricHistogram interface {
	Update(int64)
}

// MetricCounter is a simple counter
type MetricCounter interface {
	Inc(int64)
}

// Service is started and stopped in main function, which assembles services into a working application
type Service interface {
	Start() error
	Stop() error
}
