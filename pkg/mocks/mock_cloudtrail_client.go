package mocks

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/cloudtrail"
	"github.com/stretchr/testify/mock"
)

type MockCloudTrailClient struct {
	mock.Mock
}

func (m *MockCloudTrailClient) DescribeTrails(ctx context.Context, input *cloudtrail.DescribeTrailsInput, opts ...func(*cloudtrail.Options)) (*cloudtrail.DescribeTrailsOutput, error) {
	args := m.Called(ctx, input)
	if args.Get(0) != nil {
		return args.Get(0).(*cloudtrail.DescribeTrailsOutput), args.Error(1)
	}
	return nil, args.Error(1)
}
