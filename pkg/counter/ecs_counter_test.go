package counter

import (
	"aws-resource-discovery/pkg/interfaces"
	"aws-resource-discovery/pkg/mocks"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestEcsCounter_Call(t *testing.T) {
	mockClient := new(mocks.MockECSClient)
	counter := NewEcsCounter(mockClient)

	// Mock data for listClusters
	mockClient.On("ListClusters", mock.Anything, mock.Anything).Return(&ecs.ListClustersOutput{
		ClusterArns: []string{"cluster1"},
	}, nil).Once()

	// Mock data for listServices
	mockClient.On("ListServices", mock.Anything, mock.Anything).Return(&ecs.ListServicesOutput{
		ServiceArns: []string{"service1"},
	}, nil).Once()

	// Mock data for describeServices
	mockClient.On("DescribeServices", mock.Anything, mock.Anything).Return(&ecs.DescribeServicesOutput{
		Services: []types.Service{
			{
				Deployments: []types.Deployment{
					{
						TaskDefinition: aws.String("taskdef1"),
						RunningCount:   3,
					},
				},
			},
		},
	}, nil).Once()

	// Mock data for describeTaskDefinition
	mockClient.On("DescribeTaskDefinition", mock.Anything, mock.Anything).Return(&ecs.DescribeTaskDefinitionOutput{
		TaskDefinition: &types.TaskDefinition{
			ContainerDefinitions: []types.ContainerDefinition{
				{Name: aws.String("container1")},
				{Name: aws.String("container2")},
			},
		},
	}, nil).Once()

	// Test successful call
	counter.Call()

	assert.Equal(t, 6, counter.Result.Count)
	assert.Nil(t, counter.Result.Error)
	assert.Equal(t, "AWS::ECS::Cluster", counter.Result.CounterClass)

	// Test call with error
	expectedError := errors.New("test error")
	mockClient.On("ListClusters", mock.Anything, mock.Anything).Return(&ecs.ListClustersOutput{}, expectedError).Once()

	counter.Call()

	assert.Equal(t, 0, counter.Result.Count)
	assert.Equal(t, fmt.Errorf("failed to list ECS clusters: %w", expectedError), counter.Result.Error)
	assert.Equal(t, `
To scan ECS containers, the provided credentials must have the following permissions:
- ecs:ListClusters
- ecs:ListServices
- ecs:DescribeServices
- ecs:DescribeTaskDefinition
`, counter.Result.PermissionSuggestion)

	mockClient.AssertExpectations(t)
}

func TestEcsCounter_ecsCount(t *testing.T) {
	mockClient := new(mocks.MockECSClient)
	counter := NewEcsCounter(mockClient)

	// Mock data for listClusters
	mockClient.On("ListClusters", mock.Anything, mock.Anything).Return(&ecs.ListClustersOutput{
		ClusterArns: []string{"cluster1"},
	}, nil).Once()

	// Mock data for listServices
	mockClient.On("ListServices", mock.Anything, mock.Anything).Return(&ecs.ListServicesOutput{
		ServiceArns: []string{"service1"},
	}, nil).Once()

	// Mock data for describeServices
	mockClient.On("DescribeServices", mock.Anything, mock.Anything).Return(&ecs.DescribeServicesOutput{
		Services: []types.Service{
			{
				Deployments: []types.Deployment{
					{
						TaskDefinition: aws.String("taskdef1"),
						RunningCount:   2,
					},
				},
			},
		},
	}, nil).Once()

	// Mock data for describeTaskDefinition
	mockClient.On("DescribeTaskDefinition", mock.Anything, mock.Anything).Return(&ecs.DescribeTaskDefinitionOutput{
		TaskDefinition: &types.TaskDefinition{
			ContainerDefinitions: []types.ContainerDefinition{
				{Name: aws.String("container1")},
				{Name: aws.String("container2")},
			},
		},
	}, nil).Once()

	// Test ecsCount
	count, err := counter.ecsCount()
	assert.Equal(t, 4, count)
	assert.Nil(t, err)

	mockClient.AssertExpectations(t)
}

func TestEcsCounter_formatResult(t *testing.T) {
	counter := NewEcsCounter(nil)

	// Test without error
	result := counter.formatResult(10, nil)
	assert.Equal(t, 10, result.Count)
	assert.Equal(t, "AWS::ECS::Cluster", result.CounterClass)
	assert.Nil(t, result.Error)
	assert.Empty(t, result.PermissionSuggestion)

	// Test with error
	err := errors.New("test error")
	result = counter.formatResult(0, err)
	assert.Equal(t, 0, result.Count)
	assert.Equal(t, "AWS::ECS::Cluster", result.CounterClass)
	assert.Equal(t, err, result.Error)
	assert.Equal(t, "\nTo scan ECS containers, the provided credentials must have the following permissions:\n- ecs:ListClusters\n- ecs:ListServices\n- ecs:DescribeServices\n- ecs:DescribeTaskDefinition\n", result.PermissionSuggestion)
}

func TestEcsCounter_GetResult(t *testing.T) {
	counter := NewEcsCounter(nil)
	counter.Result = interfaces.CounterResult{
		Count:        5,
		CounterClass: "AWS::ECS::Cluster",
	}

	result := counter.GetResult()
	assert.Equal(t, 5, result.Count)
	assert.Equal(t, "AWS::ECS::Cluster", result.CounterClass)
}
