package mocks

import (
	"github.com/stretchr/testify/mock"
)

type MockLogger struct {
	mock.Mock
}

func (m *MockLogger) Log(record []string) error {
	args := m.Called(record)
	return args.Error(0)
}

func (m *MockLogger) Logf(format string, v ...interface{}) error {
	args := m.Called(append([]interface{}{format}, v...)...)
	return args.Error(0)
}

func (m *MockLogger) Close() error {
	args := m.Called()
	return args.Error(0)
}
