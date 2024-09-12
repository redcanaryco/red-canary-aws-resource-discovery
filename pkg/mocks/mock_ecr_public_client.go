package mocks

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/ecrpublic"
	"github.com/stretchr/testify/mock"
)

type MockECRPublicClient struct {
	mock.Mock
}

func (m *MockECRPublicClient) DescribeRepositories(ctx context.Context, input *ecrpublic.DescribeRepositoriesInput, opts ...func(*ecrpublic.Options)) (*ecrpublic.DescribeRepositoriesOutput, error) {
	args := m.Called(ctx, input)
	if output, ok := args.Get(0).(*ecrpublic.DescribeRepositoriesOutput); ok {
		return output, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockECRPublicClient) DescribeImages(ctx context.Context, input *ecrpublic.DescribeImagesInput, opts ...func(*ecrpublic.Options)) (*ecrpublic.DescribeImagesOutput, error) {
	args := m.Called(ctx, input)
	if output, ok := args.Get(0).(*ecrpublic.DescribeImagesOutput); ok {
		return output, args.Error(1)
	}
	return nil, args.Error(1)
}
