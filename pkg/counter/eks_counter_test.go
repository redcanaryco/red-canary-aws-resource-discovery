package counter

import (
	"aws-resource-discovery/pkg/interfaces"
	"aws-resource-discovery/pkg/mocks"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/eks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestEksCounter_Call(t *testing.T) {
	mockEKSClient := new(mocks.MockEKSClient)
	mockEC2Client := new(mocks.MockEC2Client)
	counter := NewEksCounter(mockEKSClient, mockEC2Client)

	// Mock data for listClusters
	mockEKSClient.On("ListClusters", mock.Anything, mock.Anything).Return(&eks.ListClustersOutput{
		Clusters: []string{"cluster1"},
	}, nil).Once()

	// Mock data for describeInstances - this may be called multiple times
	mockEC2Client.On("DescribeInstances", mock.Anything, mock.Anything).Return(&ec2.DescribeInstancesOutput{
		Reservations: []types.Reservation{
			{
				Instances: []types.Instance{
					{InstanceId: aws.String("i-1234567890abcdef0")},
				},
			},
		},
	}, nil).Once()

	// Test successful call
	counter.Call()

	assert.Equal(t, 1, counter.Result.Count)
	assert.Nil(t, counter.Result.Error)
	assert.Equal(t, "AWS::EKS::Cluster", counter.Result.CounterClass)

	// Test call with error
	expectedError := errors.New("test error")
	mockEKSClient.On("ListClusters", mock.Anything, mock.Anything).Return(&eks.ListClustersOutput{}, expectedError).Once()

	counter.Call()

	assert.Equal(t, 0, counter.Result.Count)
	assert.NotNil(t, counter.Result.Error)
	assert.Contains(t, counter.Result.Error.Error(), "failed to list EKS clusters: test error")
	assert.Equal(t, `
To scan EKS clusters, the provided credentials must have the following permissions:
- eks:ListClusters
- ec2:DescribeInstances
`, counter.Result.PermissionSuggestion)

	mockEKSClient.AssertExpectations(t)
	mockEC2Client.AssertExpectations(t)
}

func TestEksCounter_eksCount(t *testing.T) {
	mockEKSClient := new(mocks.MockEKSClient)
	mockEC2Client := new(mocks.MockEC2Client)
	counter := NewEksCounter(mockEKSClient, mockEC2Client)

	// Mock data for listClusters
	mockEKSClient.On("ListClusters", mock.Anything, mock.Anything).Return(&eks.ListClustersOutput{
		Clusters: []string{"cluster1"},
	}, nil).Once()

	// Mock data for describeInstances - adjust the number of calls as needed
	mockEC2Client.On("DescribeInstances", mock.Anything, mock.Anything).Return(&ec2.DescribeInstancesOutput{
		Reservations: []types.Reservation{
			{
				Instances: []types.Instance{
					{InstanceId: aws.String("i-1234567890abcdef0")},
				},
			},
		},
	}, nil).Once()

	// Test eksCount
	count, err := counter.eksCount()
	assert.Equal(t, 1, count)
	assert.Nil(t, err)

	mockEKSClient.AssertExpectations(t)
	mockEC2Client.AssertExpectations(t)
}

func TestEksCounter_formatResult(t *testing.T) {
	counter := NewEksCounter(nil, nil)

	// Test without error
	result := counter.formatResult(10, nil)
	assert.Equal(t, 10, result.Count)
	assert.Equal(t, "AWS::EKS::Cluster", result.CounterClass)
	assert.Nil(t, result.Error)
	assert.Empty(t, result.PermissionSuggestion)

	// Test with error
	err := errors.New("test error")
	result = counter.formatResult(0, err)
	assert.Equal(t, 0, result.Count)
	assert.Equal(t, "AWS::EKS::Cluster", result.CounterClass)
	assert.Equal(t, err, result.Error)
	assert.Equal(t, `
To scan EKS clusters, the provided credentials must have the following permissions:
- eks:ListClusters
- ec2:DescribeInstances
`, result.PermissionSuggestion)
}

func TestEksCounter_GetResult(t *testing.T) {
	counter := NewEksCounter(nil, nil)
	counter.Result = interfaces.CounterResult{
		Count:        5,
		CounterClass: "AWS::EKS::Cluster",
	}

	result := counter.GetResult()
	assert.Equal(t, 5, result.Count)
	assert.Equal(t, "AWS::EKS::Cluster", result.CounterClass)
}
