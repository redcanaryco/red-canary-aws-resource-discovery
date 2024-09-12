package mocks

import (
	"github.com/stretchr/testify/mock"
)

type MockResourceScanner struct {
	mock.Mock
}

func (m *MockResourceScanner) Call() {
	m.Called()
}