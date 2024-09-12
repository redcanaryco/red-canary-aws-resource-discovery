package interfaces

import (
	"context"

	"aws-resource-discovery/pkg/config"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/organizations/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

type STSClient interface {
	AssumeRole(ctx context.Context, input *sts.AssumeRoleInput, opts ...func(*sts.Options)) (*sts.AssumeRoleOutput, error)
}

type CredentialsManager interface {
	CredentialsFor(ctx context.Context, accountId, region string) (aws.Credentials, error)
}

type SessionManager interface {
	AssumeRole(ctx context.Context, roleArn, accountId, region string) aws.Credentials
	AssumeRoleIfNeeded(ctx context.Context, cfg aws.Config, config config.Config) aws.Credentials
	InitializeSessionAndCredentials(ctx context.Context, config config.Config, logger Logger) (aws.Config, aws.Credentials)
}
type ScanResult struct {
	Config      aws.Config
	Credentials aws.Credentials
	UserConfig  config.Config
	OrgAccounts []types.Account
}
