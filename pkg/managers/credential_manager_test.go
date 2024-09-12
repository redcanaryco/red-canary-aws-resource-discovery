package managers

import (
	"aws-resource-discovery/pkg/mocks"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/aws/aws-sdk-go-v2/service/sts/types"
	"github.com/stretchr/testify/assert"
)

func TestCredentialsManager_CredentialsFor(t *testing.T) {
	mockSTSClient := new(mocks.MockSTSClient)
	manager := NewCredentialsManager("test-role", mockSTSClient).(*credentialsManager)

	ctx := context.Background()
	accountId := "123456789012"
	region := "us-east-1"
	roleArn := "arn:aws:iam::123456789012:role/test-role"
	sessionName := "rc-aws-scan-123456789012-us-east-1"
	expiration := time.Now().Add(1 * time.Hour)

	assumeRoleInput := &sts.AssumeRoleInput{
		RoleArn:         aws.String(roleArn),
		RoleSessionName: aws.String(sessionName),
	}
	assumeRoleOutput := &sts.AssumeRoleOutput{
		Credentials: &types.Credentials{
			AccessKeyId:     aws.String("access-key-id"),
			SecretAccessKey: aws.String("secret-access-key"),
			SessionToken:    aws.String("session-token"),
			Expiration:      &expiration,
		},
	}

	mockSTSClient.On("AssumeRole", ctx, assumeRoleInput).Return(assumeRoleOutput, nil).Once()

	creds, err := manager.CredentialsFor(ctx, accountId, region)
	assert.NoError(t, err)
	assert.Equal(t, "access-key-id", creds.AccessKeyID)
	assert.Equal(t, "secret-access-key", creds.SecretAccessKey)
	assert.Equal(t, "session-token", creds.SessionToken)
	assert.Equal(t, true, creds.CanExpire)
	assert.Equal(t, expiration, creds.Expires)

	mockSTSClient.AssertExpectations(t)
}

func TestCredentialsManager_CredentialsFor_Error(t *testing.T) {
	mockSTSClient := new(mocks.MockSTSClient)
	manager := NewCredentialsManager("test-role", mockSTSClient).(*credentialsManager)

	ctx := context.TODO()
	accountId := "123456789012"
	region := "us-east-1"
	roleArn := "arn:aws:iam::123456789012:role/test-role"
	sessionName := "rc-aws-scan-123456789012-us-east-1"

	assumeRoleInput := &sts.AssumeRoleInput{
		RoleArn:         aws.String(roleArn),
		RoleSessionName: aws.String(sessionName),
	}

	expectedError := errors.New("assume role error")
	mockSTSClient.On("AssumeRole", ctx, assumeRoleInput).Return((*sts.AssumeRoleOutput)(nil), expectedError).Once()

	creds, err := manager.CredentialsFor(ctx, accountId, region)
	assert.Error(t, err)
	assert.Equal(t, "failed to assume role: assume role error", err.Error())
	assert.Equal(t, aws.Credentials{}, creds)

	mockSTSClient.AssertExpectations(t)
}

func TestCredentialsManager_createAssumeRoleInput(t *testing.T) {
	manager := NewCredentialsManager("test-role", nil).(*credentialsManager)

	accountId := "123456789012"
	region := "us-east-1"
	roleArn := "arn:aws:iam::123456789012:role/test-role"
	sessionName := "rc-aws-scan-123456789012-us-east-1"

	input := manager.createAssumeRoleInput(accountId, region)

	assert.Equal(t, roleArn, aws.ToString(input.RoleArn))
	assert.Equal(t, sessionName, aws.ToString(input.RoleSessionName))
}

func TestCredentialsManager_awsRoleArn(t *testing.T) {
	manager := NewCredentialsManager("test-role", nil).(*credentialsManager)

	accountId := "123456789012"
	expectedRoleArn := "arn:aws:iam::123456789012:role/test-role"
	roleArn := manager.awsRoleArn(accountId)

	assert.Equal(t, expectedRoleArn, roleArn)

	manager = NewCredentialsManager("", nil).(*credentialsManager)
	roleArn = manager.awsRoleArn(accountId)
	assert.Equal(t, "", roleArn)
}
