package cloudtrail

import (
	"context"
	"errors"
	"testing"

	"aws-resource-discovery/pkg/mocks"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudtrail"
	"github.com/aws/aws-sdk-go-v2/service/cloudtrail/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/stretchr/testify/assert"
)

func TestMockCloudTrailClient_DescribeTrails(t *testing.T) {
	mockClient := new(mocks.MockCloudTrailClient)
	ctx := context.Background()
	input := &cloudtrail.DescribeTrailsInput{}

	// Successful call
	expectedOutput := &cloudtrail.DescribeTrailsOutput{
		TrailList: []types.Trail{
			{TrailARN: aws.String("arn:aws:cloudtrail:us-east-1:123456789012:trail/MyTrail")},
		},
	}
	mockClient.On("DescribeTrails", ctx, input).Return(expectedOutput, nil).Once()

	output, err := mockClient.DescribeTrails(ctx, input)

	assert.NoError(t, err)
	assert.Equal(t, expectedOutput, output)
	mockClient.AssertExpectations(t)

	// Error case
	expectedError := errors.New("describe trails error")
	mockClient.On("DescribeTrails", ctx, input).Return(nil, expectedError).Once()

	output, err = mockClient.DescribeTrails(ctx, input)

	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	assert.Nil(t, output)
	mockClient.AssertExpectations(t)
}

func TestMockS3Client_GetBucketLocation(t *testing.T) {
	mockClient := new(mocks.MockS3Client)
	ctx := context.Background()
	input := &s3.GetBucketLocationInput{}

	// Successful call
	expectedOutput := &s3.GetBucketLocationOutput{
		LocationConstraint: s3types.BucketLocationConstraint("us-east-1"),
	}
	mockClient.On("GetBucketLocation", ctx, input).Return(expectedOutput, nil).Once()

	output, err := mockClient.GetBucketLocation(ctx, input)

	assert.NoError(t, err)
	assert.Equal(t, expectedOutput, output)
	mockClient.AssertExpectations(t)

	// Error case
	expectedError := errors.New("get bucket location error")
	mockClient.On("GetBucketLocation", ctx, input).Return(nil, expectedError).Once()

	output, err = mockClient.GetBucketLocation(ctx, input)

	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	assert.Nil(t, output)
	mockClient.AssertExpectations(t)
}

func TestMockS3Client_GetBucketNotificationConfiguration(t *testing.T) {
	mockClient := new(mocks.MockS3Client)
	ctx := context.Background()
	input := &s3.GetBucketNotificationConfigurationInput{}

	// Successful call
	expectedOutput := &s3.GetBucketNotificationConfigurationOutput{
		TopicConfigurations: []s3types.TopicConfiguration{
			{TopicArn: aws.String("arn:aws:sns:us-east-1:123456789012:MySNSTopic")},
		},
	}
	mockClient.On("GetBucketNotificationConfiguration", ctx, input).Return(expectedOutput, nil).Once()

	output, err := mockClient.GetBucketNotificationConfiguration(ctx, input)

	assert.NoError(t, err)
	assert.Equal(t, expectedOutput, output)
	mockClient.AssertExpectations(t)

	// Error case
	expectedError := errors.New("get bucket notification configuration error")
	mockClient.On("GetBucketNotificationConfiguration", ctx, input).Return(nil, expectedError).Once()

	output, err = mockClient.GetBucketNotificationConfiguration(ctx, input)

	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	assert.Nil(t, output)
	mockClient.AssertExpectations(t)
}
