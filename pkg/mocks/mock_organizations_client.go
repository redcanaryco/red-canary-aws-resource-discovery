package mocks

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/organizations"
	"github.com/aws/aws-sdk-go-v2/service/organizations/types"
	"github.com/stretchr/testify/mock"
)

type MockOrganizationsClient struct {
	mock.Mock
}

func (m *MockOrganizationsClient) DescribeAccount(ctx context.Context, params *organizations.DescribeAccountInput, optFns ...func(*organizations.Options)) (*organizations.DescribeAccountOutput, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*organizations.DescribeAccountOutput), args.Error(1)
}

func (m *MockOrganizationsClient) ListAccounts(ctx context.Context, params *organizations.ListAccountsInput, optFns ...func(*organizations.Options)) (*organizations.ListAccountsOutput, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*organizations.ListAccountsOutput), args.Error(1)
}

type MockOrgDetector struct {
	mock.Mock
}

func (m *MockOrgDetector) ListAccounts() []types.Account {
	args := m.Called()
	return args.Get(0).([]types.Account)
}

func (m *MockOrgDetector) SuggestPermissions(msg string) {
	m.Called(msg)
}
