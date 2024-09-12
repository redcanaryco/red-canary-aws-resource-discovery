package interfaces

import "os"

type Logger interface {
	Log(record []string) error
	Logf(format string, args ...interface{}) error
	Close() error
}

type CSVLogger interface {
	Log(record []string) error
	Logf(format string, args ...interface{}) error
	Close() error
}

type FileInitializer interface {
	InitLogger(logFile string) (*os.File, error)
	SetupLogger() (*os.File, error)
}
