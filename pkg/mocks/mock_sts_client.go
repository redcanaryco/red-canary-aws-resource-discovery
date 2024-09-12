package mocks

import (
	"aws-resource-discovery/pkg/config"
	"aws-resource-discovery/pkg/interfaces"
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/stretchr/testify/mock"
)

type MockSTSClient struct {
	mock.Mock
}

func (m *MockSTSClient) AssumeRole(ctx context.Context, params *sts.AssumeRoleInput, optFns ...func(*sts.Options)) (*sts.AssumeRoleOutput, error) {
	args := m.Called(ctx, params)
	return args.Get(0).(*sts.AssumeRoleOutput), args.Error(1)
}

type MockSessionManager struct {
	mock.Mock
}

func (m *MockSessionManager) InitializeSessionAndCredentials(ctx context.Context, config config.Config, logger interfaces.Logger) (aws.Config, aws.Credentials) {
	args := m.Called(ctx, config, logger)
	return args.Get(0).(aws.Config), args.Get(1).(aws.Credentials)
}


func (m *MockSessionManager) AssumeRole(ctx context.Context, roleArn, accountId, region string) aws.Credentials {
	args := m.Called(ctx, roleArn, accountId, region)
	return args.Get(0).(aws.Credentials)
}

func (m *MockSessionManager) AssumeRoleIfNeeded(ctx context.Context, cfg aws.Config, config config.Config) aws.Credentials {
	args := m.Called(ctx, cfg, config)
	return args.Get(0).(aws.Credentials)
}


