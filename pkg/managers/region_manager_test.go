package managers

import (
	"aws-resource-discovery/pkg/mocks"
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRegionManager_GetRegions(t *testing.T) {
	mockEC2Client := new(mocks.MockEC2Client)
	mockLogger := new(mocks.MockLogger)
	manager := NewRegionManager(mockEC2Client).(*RegionManager)

	ctx := context.TODO()
	cfg := aws.Config{}

	t.Run("Specified region", func(t *testing.T) {
		regions, err := manager.GetRegions(ctx, cfg, "us-west-1", mockLogger)
		assert.NoError(t, err)
		assert.Equal(t, []string{"us-west-1"}, regions)
	})

	t.Run("Describe regions success", func(t *testing.T) {
		mockEC2Client.On("DescribeRegions", mock.Anything, mock.Anything).Return(&ec2.DescribeRegionsOutput{
			Regions: []types.Region{
				{RegionName: aws.String("us-east-1")},
				{RegionName: aws.String("us-west-2")},
			},
		}, nil).Once()

		regions, err := manager.GetRegions(ctx, cfg, "", mockLogger)
		assert.NoError(t, err)
		assert.ElementsMatch(t, []string{"us-east-1", "us-west-2"}, regions)

		mockEC2Client.AssertExpectations(t)
	})

	t.Run("Describe regions error", func(t *testing.T) {
		expectedError := errors.New("describe regions error")
		mockEC2Client.On("DescribeRegions", mock.Anything, mock.Anything).Return(nil, expectedError).Once()
		mockLogger.On("Logf", mock.Anything, mock.Anything).Return(nil).Once()

		regions, err := manager.GetRegions(ctx, cfg, "", mockLogger)
		assert.Error(t, err)
		assert.Nil(t, regions)
		assert.Equal(t, expectedError, err)

		mockEC2Client.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})
}
