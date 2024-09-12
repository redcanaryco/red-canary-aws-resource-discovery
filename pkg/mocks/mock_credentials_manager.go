package mocks

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/stretchr/testify/mock"
)

type MockCredentialsManager struct {
	mock.Mock
}

func (m *MockCredentialsManager) CredentialsFor(ctx context.Context, accountId, region string) (aws.Credentials, error) {
	args := m.Called(ctx, accountId, region)
	return args.Get(0).(aws.Credentials), args.Error(1)
}
