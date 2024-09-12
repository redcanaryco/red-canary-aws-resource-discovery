package mocks

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/stretchr/testify/mock"
)

type MockS3Client struct {
	mock.Mock
}

func (m *MockS3Client) GetBucketLocation(ctx context.Context, input *s3.GetBucketLocationInput, opts ...func(*s3.Options)) (*s3.GetBucketLocationOutput, error) {
	args := m.Called(ctx, input)
	if args.Get(0) != nil {
		return args.Get(0).(*s3.GetBucketLocationOutput), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockS3Client) GetBucketNotificationConfiguration(ctx context.Context, input *s3.GetBucketNotificationConfigurationInput, opts ...func(*s3.Options)) (*s3.GetBucketNotificationConfigurationOutput, error) {
	args := m.Called(ctx, input)
	if args.Get(0) != nil {
		return args.Get(0).(*s3.GetBucketNotificationConfigurationOutput), args.Error(1)
	}
	return nil, args.Error(1)
}
