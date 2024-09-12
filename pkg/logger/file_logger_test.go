package logger_test

import (
	"os"
	"testing"

	"aws-resource-discovery/pkg/interfaces"
	"aws-resource-discovery/pkg/logger"

	"github.com/stretchr/testify/assert"
)

func TestInitLogger(t *testing.T) {
	filename := "test_file_logger.log"
	defer os.Remove(filename)

	logger, err := logger.InitLogger(filename)
	assert.NoError(t, err)
	assert.NotNil(t, logger)

	fileLogger, ok := logger.(interfaces.Logger)
	assert.True(t, ok)
	assert.NotNil(t, fileLogger)

	err = logger.Close()
	assert.NoError(t, err)
}

func TestFileLogger_Log(t *testing.T) {
	filename := "test_file_logger.log"
	defer os.Remove(filename)

	logger, err := logger.InitLogger(filename)
	assert.NoError(t, err)
	defer logger.Close()

	err = logger.Log([]string{"test", "data", "1"})
	assert.NoError(t, err)

	err = logger.Log([]string{"more", "test", "data"})
	assert.NoError(t, err)

	// Read the file contents to verify
	content, err := os.ReadFile(filename)
	assert.NoError(t, err)
	assert.Contains(t, string(content), "[test data 1]")
	assert.Contains(t, string(content), "[more test data]")
}

func TestFileLogger_Logf(t *testing.T) {
	filename := "test_file_logger.log"
	defer os.Remove(filename)

	logger, err := logger.InitLogger(filename)
	assert.NoError(t, err)
	defer logger.Close()

	err = logger.Logf("Formatted %s: %d", "data", 42)
	assert.NoError(t, err)

	// Read the file contents to verify
	content, err := os.ReadFile(filename)
	assert.NoError(t, err)
	assert.Contains(t, string(content), "Formatted data: 42")
}

func TestSetupLogger(t *testing.T) {
	filename := "aws-resource-discovery.log"
	defer os.Remove(filename)

	logger, err := logger.SetupLogger()
	assert.NoError(t, err)
	assert.NotNil(t, logger)

	fileLogger, ok := logger.(interfaces.Logger)
	assert.True(t, ok)
	assert.NotNil(t, fileLogger)

	err = logger.Close()
	assert.NoError(t, err)
}
