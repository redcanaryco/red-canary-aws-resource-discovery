package managers

import (
	"aws-resource-discovery/pkg/interfaces"
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

type credentialsManager struct {
	awsRoleName string
	stsClient   interfaces.STSClient
	profile     string
	debug       bool
}

func NewCredentialsManager(roleName string, stsClient interfaces.STSClient, profile string, debug bool) interfaces.CredentialsManager {
	return &credentialsManager{
		awsRoleName: roleName,
		stsClient:   stsClient,
		profile:     profile,
		debug:       debug,
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
	if cm.profile != "" {
		if cm.debug {
			fmt.Println("\n[DEBUG] Using AWS profile:", cm.profile)
		}
		if cm.debug {
			cfg, err := config.LoadDefaultConfig(ctx, config.WithSharedConfigProfile(cm.profile))
			if err == nil {
				creds, err := cfg.Credentials.Retrieve(ctx)
				if err == nil {
					fmt.Printf("[DEBUG] SDK loaded access key: %s\n", creds.AccessKeyID)
				} else {
					fmt.Printf("[DEBUG] Could not retrieve credentials from SDK: %v\n", err)
				}
			} else {
				fmt.Printf("[DEBUG] Could not load AWS SDK config: %v\n", err)
			}
		}
		return aws.Credentials{}, nil
	}
	if cm.debug {
		fmt.Println("\n[DEBUG] No profile set, attempting to AssumeRole via STS")
	}
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
