package zero_logger_adapter

type ILogger interface {
	Write(p []byte) (n int, err error)
	Close() error
}
