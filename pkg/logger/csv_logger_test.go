package logger_test

import (
	"os"
	"testing"

	"aws-resource-discovery/pkg/interfaces"
	"aws-resource-discovery/pkg/logger"

	"github.com/stretchr/testify/assert"
)

func TestNewCSVLogger(t *testing.T) {
	filename := "test_csv_logger.csv"
	defer os.Remove(filename)

	logger, err := logger.NewCSVLogger(filename)
	assert.NoError(t, err)
	assert.NotNil(t, logger)

	csvLogger, ok := logger.(interfaces.CSVLogger)
	assert.True(t, ok)
	assert.NotNil(t, csvLogger)

	err = logger.Close()
	assert.NoError(t, err)
}

func TestCSVLogger_Log(t *testing.T) {
	filename := "test_csv_logger.csv"
	defer os.Remove(filename)

	logger, err := logger.NewCSVLogger(filename)
	assert.NoError(t, err)
	defer logger.Close()

	err = logger.Log([]string{"test", "data", "1"})
	assert.NoError(t, err)

	err = logger.Log([]string{"more", "test", "data"})
	assert.NoError(t, err)

	// Read the file contents to verify
	content, err := os.ReadFile(filename)
	assert.NoError(t, err)
	expectedContent := "test,data,1\nmore,test,data\n"
	assert.Equal(t, expectedContent, string(content))
}

func TestCSVLogger_Logf(t *testing.T) {
	filename := "test_csv_logger.csv"
	defer os.Remove(filename)

	logger, err := logger.NewCSVLogger(filename)
	assert.NoError(t, err)
	defer logger.Close()

	err = logger.Logf("Formatted %s: %d", "data", 42)
	assert.NoError(t, err)

	// Read the file contents to verify
	content, err := os.ReadFile(filename)
	assert.NoError(t, err)
	expectedContent := "Formatted data: 42\n"
	assert.Equal(t, expectedContent, string(content))
}
