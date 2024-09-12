package logger

import (
	"aws-resource-discovery/pkg/interfaces"
	"log"
	"os"
)

type fileLogger struct {
	logger *log.Logger
	file   *os.File
}

func InitLogger(logFile string) (interfaces.Logger, error) {
	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}
	logger := log.New(file, "", log.LstdFlags)
	return &fileLogger{logger: logger, file: file}, nil
}

func (l *fileLogger) Log(record []string) error {
	l.logger.Println(record)
	return nil
}

func (l *fileLogger) Logf(format string, args ...interface{}) error {
	l.logger.Printf(format, args...)
	return nil
}

func (l *fileLogger) Close() error {
	return l.file.Close()
}

func SetupLogger() (interfaces.Logger, error) {
	logFileName := "aws-resource-discovery.log"

	if _, err := os.Stat(logFileName); err == nil {
		err := os.Remove(logFileName)
		if err != nil {
			log.Fatalf("Failed to remove existing logger file: %v", err)
		}
	}

	return InitLogger(logFileName)
}
