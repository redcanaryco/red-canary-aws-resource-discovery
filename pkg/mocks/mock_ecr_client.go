package mocks

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/ecr"
	"github.com/stretchr/testify/mock"
)

type MockECRClient struct {
	mock.Mock
}

func (m *MockECRClient) DescribeRepositories(ctx context.Context, input *ecr.DescribeRepositoriesInput, opts ...func(*ecr.Options)) (*ecr.DescribeRepositoriesOutput, error) {
	args := m.Called(ctx, input)
	if output, ok := args.Get(0).(*ecr.DescribeRepositoriesOutput); ok {
		return output, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockECRClient) ListImages(ctx context.Context, input *ecr.ListImagesInput, opts ...func(*ecr.Options)) (*ecr.ListImagesOutput, error) {
	args := m.Called(ctx, input)
	if output, ok := args.Get(0).(*ecr.ListImagesOutput); ok {
		return output, args.Error(1)
	}
	return nil, args.Error(1)
}
