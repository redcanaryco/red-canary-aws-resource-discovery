package logger

import (
	"aws-resource-discovery/pkg/interfaces"
	"encoding/csv"
	"fmt"
	"os"
)

type csvLogger struct {
	file   *os.File
	writer *csv.Writer
}

func NewCSVLogger(filename string) (interfaces.Logger, error) {
	file, err := os.Create(filename)
	if err != nil {
		return nil, err
	}

	writer := csv.NewWriter(file)
	return &csvLogger{file: file, writer: writer}, nil
}

func (l *csvLogger) Log(record []string) error {
	err := l.writer.Write(record)
	if err != nil {
		return err
	}
	l.writer.Flush()
	return nil
}

func (l *csvLogger) Logf(format string, args ...interface{}) error {
	record := fmt.Sprintf(format, args...)
	return l.Log([]string{record})
}

func (l *csvLogger) Close() error {
	l.writer.Flush()
	return l.file.Close()
}
