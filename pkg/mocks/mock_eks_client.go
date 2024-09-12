package mocks

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/eks"
	"github.com/stretchr/testify/mock"
)

type MockEKSClient struct {
	mock.Mock
}

func (m *MockEKSClient) ListClusters(ctx context.Context, params *eks.ListClustersInput, optFns ...func(*eks.Options)) (*eks.ListClustersOutput, error) {
	args := m.Called(ctx, params)
	return args.Get(0).(*eks.ListClustersOutput), args.Error(1)
}

func (m *MockEKSClient) DescribeCluster(ctx context.Context, params *eks.DescribeClusterInput, optFns ...func(*eks.Options)) (*eks.DescribeClusterOutput, error) {
	args := m.Called(ctx, params)
	return args.Get(0).(*eks.DescribeClusterOutput), args.Error(1)
}
