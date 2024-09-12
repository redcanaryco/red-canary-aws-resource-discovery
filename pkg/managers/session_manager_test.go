package managers

import (
	"aws-resource-discovery/pkg/config"
	"aws-resource-discovery/pkg/mocks"
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/aws/aws-sdk-go-v2/service/sts/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSessionManager_AssumeRole(t *testing.T) {
	mockClient := new(mocks.MockSTSClient)
	manager := NewSessionManager(mockClient)

	ctx := context.TODO()
	roleArn := "arn:aws:iam::123456789012:role/test-role"
	accountId := "123456789012"
	region := "us-east-1"
	expiration := time.Now().Add(1 * time.Hour)

	mockClient.On("AssumeRole", ctx, mock.AnythingOfType("*sts.AssumeRoleInput")).Return(&sts.AssumeRoleOutput{
		Credentials: &types.Credentials{
			AccessKeyId:     aws.String("accessKeyId"),
			SecretAccessKey: aws.String("secretAccessKey"),
			SessionToken:    aws.String("sessionToken"),
			Expiration:      &expiration,
		},
	}, nil).Once()

	creds := manager.AssumeRole(ctx, roleArn, accountId, region)

	assert.Equal(t, "accessKeyId", creds.AccessKeyID)
	assert.Equal(t, "secretAccessKey", creds.SecretAccessKey)
	assert.Equal(t, "sessionToken", creds.SessionToken)
	assert.True(t, creds.CanExpire)
	assert.Equal(t, expiration, creds.Expires)

	mockClient.AssertExpectations(t)
}

func TestSessionManager_AssumeRoleIfNeeded(t *testing.T) {
	mockClient := new(mocks.MockSTSClient)
	manager := NewSessionManager(mockClient)

	ctx := context.TODO()
	cfg := aws.Config{}
	accountId := "123456789012"
	region := "us-east-1"
	roleArn := "arn:aws:iam::123456789012:role/test-role"
	expiration := time.Now().Add(1 * time.Hour)

	t.Run("No RoleArn provided", func(t *testing.T) {
		conf := config.Config{
			AccountId: accountId,
			Region:    region,
		}
		creds := manager.AssumeRoleIfNeeded(ctx, cfg, conf)
		assert.Empty(t, creds.AccessKeyID)
		assert.Empty(t, creds.SecretAccessKey)
		assert.Empty(t, creds.SessionToken)
	})

	t.Run("RoleArn provided", func(t *testing.T) {
		conf := config.Config{
			AccountId: accountId,
			Region:    region,
			RoleArn:   roleArn,
		}

		mockClient.On("AssumeRole", ctx, mock.AnythingOfType("*sts.AssumeRoleInput")).Return(&sts.AssumeRoleOutput{
			Credentials: &types.Credentials{
				AccessKeyId:     aws.String("accessKeyId"),
				SecretAccessKey: aws.String("secretAccessKey"),
				SessionToken:    aws.String("sessionToken"),
				Expiration:      &expiration,
			},
		}, nil).Once()

		creds := manager.AssumeRoleIfNeeded(ctx, cfg, conf)
		assert.Equal(t, "accessKeyId", creds.AccessKeyID)
		assert.Equal(t, "secretAccessKey", creds.SecretAccessKey)
		assert.Equal(t, "sessionToken", creds.SessionToken)
		assert.True(t, creds.CanExpire)
		assert.Equal(t, expiration, creds.Expires)

		mockClient.AssertExpectations(t)
	})
}

func TestSessionManager_InitializeSessionAndCredentials(t *testing.T) {
	mockClient := new(mocks.MockSTSClient)
	mockLogger := new(mocks.MockLogger)
	manager := NewSessionManager(mockClient)

	ctx := context.TODO()
	accountId := "123456789012"
	region := "us-east-1"
	roleArn := "arn:aws:iam::123456789012:role/test-role"
	expiration := time.Now().Add(1 * time.Hour)

	mockClient.On("AssumeRole", ctx, mock.AnythingOfType("*sts.AssumeRoleInput")).Return(&sts.AssumeRoleOutput{
		Credentials: &types.Credentials{
			AccessKeyId:     aws.String("accessKeyId"),
			SecretAccessKey: aws.String("secretAccessKey"),
			SessionToken:    aws.String("sessionToken"),
			Expiration:      &expiration,
		},
	}, nil).Once()

	conf := config.Config{
		AccountId: accountId,
		Region:    region,
		RoleArn:   roleArn,
	}

	awsCfg, creds := manager.InitializeSessionAndCredentials(ctx, conf, mockLogger)
	assert.NotNil(t, awsCfg)
	assert.Equal(t, "accessKeyId", creds.AccessKeyID)
	assert.Equal(t, "secretAccessKey", creds.SecretAccessKey)
	assert.Equal(t, "sessionToken", creds.SessionToken)
	assert.True(t, creds.CanExpire)
	assert.Equal(t, expiration, creds.Expires)

	mockClient.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}
