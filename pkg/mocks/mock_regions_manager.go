package mocks

import (
	"aws-resource-discovery/pkg/interfaces"
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/stretchr/testify/mock"
)

type MockRegionsManager struct {
	mock.Mock
}

func (m *MockRegionsManager) GetRegions(ctx context.Context, cfg aws.Config, specifiedRegion string, logger interfaces.Logger) ([]string, error) {
	args := m.Called(ctx, cfg, specifiedRegion, logger)
	return args.Get(0).([]string), args.Error(1)
}
