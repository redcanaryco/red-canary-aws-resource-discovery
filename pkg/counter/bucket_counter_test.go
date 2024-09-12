package counter

import (
	"aws-resource-discovery/pkg/interfaces"
	"aws-resource-discovery/pkg/mocks"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/cloudcontrol"
	"github.com/aws/aws-sdk-go-v2/service/cloudcontrol/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestBucketCounter_Call(t *testing.T) {
	// Test successful call
	mockClient := new(mocks.MockCloudControlClient)
	counter := NewBucketCounter(mockClient)
	mockClient.On("ListResources", mock.Anything, mock.Anything).Return(&cloudcontrol.ListResourcesOutput{
		ResourceDescriptions: make([]types.ResourceDescription, 5),
	}, nil).Once()

	counter.Call()

	assert.Equal(t, 5, counter.Result.Count, "Expected count to be 5")
	assert.Nil(t, counter.Result.Error, "Expected error to be nil")
	assert.Equal(t, "AWS::S3::Bucket", counter.Result.CounterClass, "Expected counter class to be AWS::S3::Bucket")

	// Test call with error
	mockClient = new(mocks.MockCloudControlClient)
	counter = NewBucketCounter(mockClient)
	expectedError := errors.New("test error")
	mockClient.On("ListResources", mock.Anything, mock.Anything).Return(&cloudcontrol.ListResourcesOutput{}, expectedError).Once()

	counter.Call()

	assert.Equal(t, 0, counter.Result.Count, "Expected count to be 0")
	assert.Equal(t, expectedError, counter.Result.Error, "Expected error to be test error")
	assert.Equal(t, "\nTo scan S3 buckets, the provided credentials must have the following permissions:\n- s3:ListAllMyBuckets\n- s3:GetBucketLocation\n", counter.Result.PermissionSuggestion, "Expected permission suggestion to be specific string")

	mockClient.AssertExpectations(t)
}

func TestBucketCounter_paginatedCount(t *testing.T) {
	mockClient := new(mocks.MockCloudControlClient)
	counter := NewBucketCounter(mockClient)

	// Test single page
	mockClient.On("ListResources", mock.Anything, mock.Anything).Return(&cloudcontrol.ListResourcesOutput{
		ResourceDescriptions: make([]types.ResourceDescription, 5),
	}, nil).Once()

	count := counter.paginatedCount()
	assert.Equal(t, 5, count)

	// Test multiple pages
	mockClient.On("ListResources", mock.Anything, mock.Anything).Return(&cloudcontrol.ListResourcesOutput{
		ResourceDescriptions: make([]types.ResourceDescription, 5),
		NextToken:            new(string),
	}, nil).Once()
	mockClient.On("ListResources", mock.Anything, mock.Anything).Return(&cloudcontrol.ListResourcesOutput{
		ResourceDescriptions: make([]types.ResourceDescription, 3),
	}, nil).Once()

	count = counter.paginatedCount()
	assert.Equal(t, 8, count)

	mockClient.AssertExpectations(t)
}

func TestBucketCounter_formatResult(t *testing.T) {
	counter := NewBucketCounter(nil)

	// Test without error
	result := counter.formatResult(10, nil)
	assert.Equal(t, 10, result.Count)
	assert.Equal(t, "AWS::S3::Bucket", result.CounterClass)
	assert.Nil(t, result.Error)
	assert.Empty(t, result.PermissionSuggestion)

	// Test with error
	err := errors.New("test error")
	result = counter.formatResult(0, err)
	assert.Equal(t, 0, result.Count)
	assert.Equal(t, "AWS::S3::Bucket", result.CounterClass)
	assert.Equal(t, err, result.Error)
	assert.Equal(t, "\nTo scan S3 buckets, the provided credentials must have the following permissions:\n- s3:ListAllMyBuckets\n- s3:GetBucketLocation\n", result.PermissionSuggestion)
}

func TestBucketCounter_GetResult(t *testing.T) {
	counter := NewBucketCounter(nil)
	counter.Result = interfaces.CounterResult{
		Count:        5,
		CounterClass: "AWS::S3::Bucket",
	}

	result := counter.GetResult()
	assert.Equal(t, 5, result.Count)
	assert.Equal(t, "AWS::S3::Bucket", result.CounterClass)
}
