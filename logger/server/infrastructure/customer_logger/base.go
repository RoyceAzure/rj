package customerlogger

type ILogger interface {
	Write(p []byte) (n int, err error)
	Close() error
}
