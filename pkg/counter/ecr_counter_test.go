package counter

import (
	"aws-resource-discovery/pkg/interfaces"
	"aws-resource-discovery/pkg/mocks"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
	"github.com/aws/aws-sdk-go-v2/service/ecr/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestEcrCounter_Call(t *testing.T) {
	mockClient := new(mocks.MockECRClient)
	counter := NewEcrCounter(mockClient)

	// Mock data for DescribeRepositories
	mockClient.On("DescribeRepositories", mock.Anything, mock.Anything).Return(&ecr.DescribeRepositoriesOutput{
		Repositories: []types.Repository{
			{RepositoryName: aws.String("repo1")},
			{RepositoryName: aws.String("repo2")},
		},
	}, nil).Once()

	// Mock data for ListImages
	mockClient.On("ListImages", mock.Anything, mock.Anything).Return(&ecr.ListImagesOutput{
		ImageIds: []types.ImageIdentifier{
			{ImageTag: aws.String("image1")},
			{ImageTag: aws.String("image2")},
		},
	}, nil).Twice()

	// Test successful call
	counter.Call()

	assert.Equal(t, 4, counter.Result.Count)
	assert.Nil(t, counter.Result.Error)
	assert.Equal(t, "AWS::ECR::Repository", counter.Result.CounterClass)

	// Test call with error
	expectedError := errors.New("test error")
	mockClient.On("DescribeRepositories", mock.Anything, mock.Anything).Return(&ecr.DescribeRepositoriesOutput{}, expectedError).Once()

	counter.Call()

	assert.Equal(t, 0, counter.Result.Count)
	assert.NotNil(t, counter.Result.Error)
	assert.Contains(t, counter.Result.Error.Error(), "failed to list ECR repositories: test error")

	assert.Equal(t, `
To scan ECR repositories, the provided credentials must have the following permissions:
  - ecr:DescribeRepositories
  - ecr:ListImages
`, counter.Result.PermissionSuggestion)

	mockClient.AssertExpectations(t)
}

func TestEcrCounter_ecrCount(t *testing.T) {
	mockClient := new(mocks.MockECRClient)
	counter := NewEcrCounter(mockClient)

	// Mock data for DescribeRepositories
	mockClient.On("DescribeRepositories", mock.Anything, mock.Anything).Return(&ecr.DescribeRepositoriesOutput{
		Repositories: []types.Repository{
			{RepositoryName: aws.String("repo1")},
		},
	}, nil).Once()

	// Mock data for ListImages
	mockClient.On("ListImages", mock.Anything, mock.Anything).Return(&ecr.ListImagesOutput{
		ImageIds: []types.ImageIdentifier{
			{ImageTag: aws.String("image1")},
			{ImageTag: aws.String("image2")},
		},
	}, nil).Once()

	// Test ecrCount
	count, err := counter.ecrCount()
	assert.Equal(t, 2, count)
	assert.Nil(t, err)

	mockClient.AssertExpectations(t)
}

func TestEcrCounter_countImagesInRepository(t *testing.T) {
	mockClient := new(mocks.MockECRClient)
	counter := NewEcrCounter(mockClient)

	// Mock data for ListImages
	mockClient.On("ListImages", mock.Anything, mock.Anything).Return(&ecr.ListImagesOutput{
		ImageIds: []types.ImageIdentifier{
			{ImageTag: aws.String("image1")},
			{ImageTag: aws.String("image2")},
		},
	}, nil).Once()

	// Test countImagesInRepository
	count, err := counter.countImagesInRepository(aws.String("repo1"))
	assert.Equal(t, 2, count)
	assert.Nil(t, err)

	mockClient.AssertExpectations(t)
}

func TestEcrCounter_formatResult(t *testing.T) {
	counter := NewEcrCounter(nil)

	// Test without error
	result := counter.formatResult(10, nil)
	assert.Equal(t, 10, result.Count)
	assert.Equal(t, "AWS::ECR::Repository", result.CounterClass)
	assert.Nil(t, result.Error)
	assert.Empty(t, result.PermissionSuggestion)

	// Test with error
	err := errors.New("test error")
	result = counter.formatResult(0, err)
	assert.Equal(t, 0, result.Count)
	assert.Equal(t, "AWS::ECR::Repository", result.CounterClass)
	assert.Equal(t, err, result.Error)
	assert.Equal(t, `
To scan ECR repositories, the provided credentials must have the following permissions:
  - ecr:DescribeRepositories
  - ecr:ListImages
`, result.PermissionSuggestion)
}

func TestEcrCounter_GetResult(t *testing.T) {
	counter := NewEcrCounter(nil)
	counter.Result = interfaces.CounterResult{
		Count:        5,
		CounterClass: "AWS::ECR::Repository",
	}

	result := counter.GetResult()
	assert.Equal(t, 5, result.Count)
	assert.Equal(t, "AWS::ECR::Repository", result.CounterClass)
}
