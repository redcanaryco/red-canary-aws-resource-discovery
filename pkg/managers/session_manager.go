package managers

import (
	"aws-resource-discovery/pkg/config"
	"aws-resource-discovery/pkg/interfaces"
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"

	aws_conf "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

type sessionManager struct {
	Client interfaces.STSClient
}

func NewSessionManager(client interfaces.STSClient) interfaces.SessionManager {
	return &sessionManager{
		Client: client,
	}
}

func (sm *sessionManager) AssumeRole(ctx context.Context, roleArn, accountId, region string) aws.Credentials {
	input := &sts.AssumeRoleInput{
		RoleArn:         aws.String(roleArn),
		RoleSessionName: aws.String(fmt.Sprintf("rc-aws-scan-%s-%s", accountId, region)),
	}

	result, err := sm.Client.AssumeRole(ctx, input)
	if err != nil {
		log.Fatalf("Failed to assume role: %v", err)
	}
	return aws.Credentials{
		AccessKeyID:     aws.ToString(result.Credentials.AccessKeyId),
		SecretAccessKey: aws.ToString(result.Credentials.SecretAccessKey),
		SessionToken:    aws.ToString(result.Credentials.SessionToken),
		CanExpire:       true,
		Expires:         *result.Credentials.Expiration,
	}
}

func (sm *sessionManager) AssumeRoleIfNeeded(ctx context.Context, cfg aws.Config, config config.Config) aws.Credentials {
	if config.RoleArn == "" {
		return aws.Credentials{}
	}
	input := &sts.AssumeRoleInput{
		RoleArn:         aws.String(config.RoleArn),
		RoleSessionName: aws.String(fmt.Sprintf("rc-aws-scan-%s-%s", config.AccountId, config.Region)),
	}

	assumeRoleOutput, err := sm.Client.AssumeRole(ctx, input)
	if err != nil {
		log.Fatalf("Failed to assume role: %v", err)
	}
	return aws.Credentials{
		AccessKeyID:     aws.ToString(assumeRoleOutput.Credentials.AccessKeyId),
		SecretAccessKey: aws.ToString(assumeRoleOutput.Credentials.SecretAccessKey),
		SessionToken:    aws.ToString(assumeRoleOutput.Credentials.SessionToken),
		CanExpire:       true,
		Expires:         *assumeRoleOutput.Credentials.Expiration,
	}
}

func (sm *sessionManager) InitializeSessionAndCredentials(ctx context.Context, config config.Config, logger interfaces.Logger) (aws.Config, aws.Credentials) {
	cfg, err := aws_conf.LoadDefaultConfig(ctx, aws_conf.WithRegion(config.Region))
	if err != nil {
		logger.Logf("Failed to load AWS config: %v", err)
		return aws.Config{}, aws.Credentials{}
	}

	credentials := sm.AssumeRoleIfNeeded(ctx, cfg, config)
	return cfg, credentials
}
