package managers

import (
	"aws-resource-discovery/pkg/interfaces"
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

type credentialsManager struct {
	awsRoleName string
	stsClient   interfaces.STSClient
}

func NewCredentialsManager(roleName string, stsClient interfaces.STSClient) interfaces.CredentialsManager {
	return &credentialsManager{
		awsRoleName: roleName,
		stsClient:   stsClient,
	}
}

func (cm *credentialsManager) createAssumeRoleInput(accountId, region string) *sts.AssumeRoleInput {
	roleArn := cm.awsRoleArn(accountId)
	if roleArn == "" {
		roleArn = fmt.Sprintf("arn:aws:iam::%s:role/red-canary-resource-discovery-role", accountId)
	}

	input := &sts.AssumeRoleInput{
		RoleArn:         aws.String(roleArn),
		RoleSessionName: aws.String(fmt.Sprintf("rc-aws-scan-%s-%s", accountId, region)),
	}

	return input
}

func (cm *credentialsManager) CredentialsFor(ctx context.Context, accountId, region string) (aws.Credentials, error) {
	input := cm.createAssumeRoleInput(accountId, region)

	assumeRoleOutput, err := cm.stsClient.AssumeRole(ctx, input)
	if err != nil {
		return aws.Credentials{}, fmt.Errorf("failed to assume role: %w", err)
	}
	return aws.Credentials{
		AccessKeyID:     aws.ToString(assumeRoleOutput.Credentials.AccessKeyId),
		SecretAccessKey: aws.ToString(assumeRoleOutput.Credentials.SecretAccessKey),
		SessionToken:    aws.ToString(assumeRoleOutput.Credentials.SessionToken),
		CanExpire:       true,
		Expires:         *assumeRoleOutput.Credentials.Expiration,
	}, nil
}

func (cm *credentialsManager) awsRoleArn(accountId string) string {
	if cm.awsRoleName == "" {
		return ""
	}
	return fmt.Sprintf("arn:aws:iam::%s:role/%s", accountId, cm.awsRoleName)
}
