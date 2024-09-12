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

func TestBaseCounter_Call(t *testing.T) {
	mockClient := new(mocks.MockCloudControlClient)
	counter := &BaseCounter{
		Client:   mockClient,
		TypeName: "TestResource",
		PermissionSuggestionFunc: func() string {
			return "Test permission suggestion"
		},
	}

	// Test successful call
	mockClient.On("ListResources", mock.Anything, mock.Anything, mock.Anything).Return(&cloudcontrol.ListResourcesOutput{
		ResourceDescriptions: make([]types.ResourceDescription, 5),
	}, nil).Once()

	counter.Call()

	assert.Equal(t, 5, counter.Result.Count)
	assert.Nil(t, counter.Result.Error)
	assert.Equal(t, "TestResource", counter.Result.CounterClass)

	// Test call with error
	expectedError := errors.New("test error")
	mockClient.On("ListResources", mock.Anything, mock.Anything, mock.Anything).Return(&cloudcontrol.ListResourcesOutput{}, expectedError).Once()

	counter.Call()

	assert.Equal(t, 0, counter.Result.Count)
	assert.Equal(t, expectedError, counter.Result.Error)
	assert.Equal(t, "Test permission suggestion", counter.Result.PermissionSuggestion)

	mockClient.AssertExpectations(t)
}

func TestBaseCounter_paginatedCount(t *testing.T) {
	mockClient := new(mocks.MockCloudControlClient)
	counter := &BaseCounter{
		Client:   mockClient,
		TypeName: "TestResource",
	}

	// Test single page
	mockClient.On("ListResources", mock.Anything, mock.Anything, mock.Anything).Return(&cloudcontrol.ListResourcesOutput{
		ResourceDescriptions: make([]types.ResourceDescription, 5),
	}, nil).Once()

	count := counter.paginatedCount()
	assert.Equal(t, 5, count)

	// Test multiple pages
	mockClient.On("ListResources", mock.Anything, mock.Anything, mock.Anything).Return(&cloudcontrol.ListResourcesOutput{
		ResourceDescriptions: make([]types.ResourceDescription, 5),
		NextToken:            new(string),
	}, nil).Once()
	mockClient.On("ListResources", mock.Anything, mock.Anything, mock.Anything).Return(&cloudcontrol.ListResourcesOutput{
		ResourceDescriptions: make([]types.ResourceDescription, 3),
	}, nil).Once()

	count = counter.paginatedCount()
	assert.Equal(t, 8, count)

	mockClient.AssertExpectations(t)
}

func TestBaseCounter_formatResult(t *testing.T) {
	counter := &BaseCounter{
		TypeName: "TestResource",
		PermissionSuggestionFunc: func() string {
			return "Test permission suggestion"
		},
	}

	// Test without error
	result := counter.formatResult(10, nil)
	assert.Equal(t, 10, result.Count)
	assert.Equal(t, "TestResource", result.CounterClass)
	assert.Nil(t, result.Error)
	assert.Empty(t, result.PermissionSuggestion)

	// Test with error
	err := errors.New("test error")
	result = counter.formatResult(0, err)
	assert.Equal(t, 0, result.Count)
	assert.Equal(t, "TestResource", result.CounterClass)
	assert.Equal(t, err, result.Error)
	assert.Equal(t, "Test permission suggestion", result.PermissionSuggestion)
}

func TestBaseCounter_GetResult(t *testing.T) {
	counter := &BaseCounter{
		Result: interfaces.CounterResult{
			Count:        5,
			CounterClass: "TestResource",
		},
	}

	result := counter.GetResult()
	assert.Equal(t, 5, result.Count)
	assert.Equal(t, "TestResource", result.CounterClass)
}
