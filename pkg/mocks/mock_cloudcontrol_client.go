package mocks

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/cloudcontrol"
	"github.com/stretchr/testify/mock"
)

type MockCloudControlClient struct {
	mock.Mock
}

func (m *MockCloudControlClient) ListResources(ctx context.Context, params *cloudcontrol.ListResourcesInput, optFns ...func(*cloudcontrol.Options)) (*cloudcontrol.ListResourcesOutput, error) {
	args := m.Called(ctx, params)
	return args.Get(0).(*cloudcontrol.ListResourcesOutput), args.Error(1)
}
