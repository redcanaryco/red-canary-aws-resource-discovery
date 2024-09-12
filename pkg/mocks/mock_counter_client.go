package mocks

import (
	"aws-resource-discovery/pkg/interfaces"

	"github.com/stretchr/testify/mock"
)

type MockCounter struct {
	mock.Mock
}

func (m *MockCounter) Call() {
	m.Called()
}

func (m *MockCounter) GetResult() interfaces.CounterResult {
	args := m.Called()
	return args.Get(0).(interfaces.CounterResult)
}
