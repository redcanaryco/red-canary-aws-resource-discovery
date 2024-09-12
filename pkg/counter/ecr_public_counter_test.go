package counter

import (
	"aws-resource-discovery/pkg/interfaces"
	"aws-resource-discovery/pkg/mocks"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecrpublic"
	"github.com/aws/aws-sdk-go-v2/service/ecrpublic/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestEcrPublicCounter_Call(t *testing.T) {
	mockClient := new(mocks.MockECRPublicClient)
	counter := NewEcrPublicCounter(mockClient)

	// Mock data for DescribeRepositories
	mockClient.On("DescribeRepositories", mock.Anything, mock.Anything).Return(&ecrpublic.DescribeRepositoriesOutput{
		Repositories: []types.Repository{
			{RepositoryName: aws.String("repo1")},
			{RepositoryName: aws.String("repo2")},
		},
	}, nil).Once()

	// Mock data for DescribeImages
	mockClient.On("DescribeImages", mock.Anything, mock.Anything).Return(&ecrpublic.DescribeImagesOutput{
		ImageDetails: []types.ImageDetail{
			{ImageDigest: aws.String("digest1")},
			{ImageDigest: aws.String("digest2")},
		},
	}, nil).Twice() // Called twice for two repositories

	// Test successful call
	counter.Call()

	assert.Equal(t, 4, counter.Result.Count)
	assert.Nil(t, counter.Result.Error)
	assert.Equal(t, "AWS::ECR::PublicRepository", counter.Result.CounterClass)

	// Test call with error
	expectedError := errors.New("test error")
	mockClient.On("DescribeRepositories", mock.Anything, mock.Anything).Return(&ecrpublic.DescribeRepositoriesOutput{}, expectedError).Once()

	counter.Call()

	assert.Equal(t, 0, counter.Result.Count)
	assert.NotNil(t, counter.Result.Error)
	assert.Contains(t, counter.Result.Error.Error(), "failed to list public ECR repositories: test error")
	assert.Equal(t, `
To scan public ECR repositories, the provided credentials must have the following permissions:
  - ecr-public:DescribeRepositories
  - ecr-public:DescribeImages
`, counter.Result.PermissionSuggestion)

	mockClient.AssertExpectations(t)
}

func TestEcrPublicCounter_ecrPublicCount(t *testing.T) {
	mockClient := new(mocks.MockECRPublicClient)
	counter := NewEcrPublicCounter(mockClient)

	// Mock data for DescribeRepositories
	mockClient.On("DescribeRepositories", mock.Anything, mock.Anything).Return(&ecrpublic.DescribeRepositoriesOutput{
		Repositories: []types.Repository{
			{RepositoryName: aws.String("repo1")},
		},
	}, nil).Once()

	// Mock data for DescribeImages
	mockClient.On("DescribeImages", mock.Anything, mock.Anything).Return(&ecrpublic.DescribeImagesOutput{
		ImageDetails: []types.ImageDetail{
			{ImageDigest: aws.String("digest1")},
			{ImageDigest: aws.String("digest2")},
		},
	}, nil).Once()

	// Test ecrPublicCount
	count, err := counter.ecrPublicCount()
	assert.Equal(t, 2, count)
	assert.Nil(t, err)

	mockClient.AssertExpectations(t)
}

func TestEcrPublicCounter_formatResult(t *testing.T) {
	counter := NewEcrPublicCounter(nil)

	// Test without error
	result := counter.formatResult(10, nil)
	assert.Equal(t, 10, result.Count)
	assert.Equal(t, "AWS::ECR::PublicRepository", result.CounterClass)
	assert.Nil(t, result.Error)
	assert.Empty(t, result.PermissionSuggestion)

	// Test with error
	err := errors.New("test error")
	result = counter.formatResult(0, err)
	assert.Equal(t, 0, result.Count)
	assert.Equal(t, "AWS::ECR::PublicRepository", result.CounterClass)
	assert.Equal(t, err, result.Error)
	assert.Equal(t, `
To scan public ECR repositories, the provided credentials must have the following permissions:
  - ecr-public:DescribeRepositories
  - ecr-public:DescribeImages
`, result.PermissionSuggestion)
}

func TestEcrPublicCounter_GetResult(t *testing.T) {
	counter := NewEcrPublicCounter(nil)
	counter.Result = interfaces.CounterResult{
		Count:        5,
		CounterClass: "AWS::ECR::PublicRepository",
	}

	result := counter.GetResult()
	assert.Equal(t, 5, result.Count)
	assert.Equal(t, "AWS::ECR::PublicRepository", result.CounterClass)
}
