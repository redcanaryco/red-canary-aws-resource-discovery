package utils

import (
	"aws-resource-discovery/pkg/mocks"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/organizations"
	"github.com/aws/aws-sdk-go-v2/service/organizations/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestFilterActiveAccounts(t *testing.T) {
	mockOrgClient := new(mocks.MockOrganizationsClient)
	mockLogger := new(mocks.MockLogger)

	accounts := []types.Account{
		{Id: aws.String("111111111111"), Status: types.AccountStatusActive},
		{Id: aws.String("222222222222"), Status: types.AccountStatusSuspended},
		{Id: aws.String("333333333333"), Status: types.AccountStatusActive},
	}

	excludeAccounts := []string{"333333333333"}

	filter := NewAccountFilter(accounts, mockOrgClient, mockLogger, excludeAccounts)

	mockOrgClient.On("DescribeAccount", mock.Anything, &organizations.DescribeAccountInput{
		AccountId: aws.String("111111111111"),
	}).Return(&organizations.DescribeAccountOutput{
		Account: &types.Account{Status: types.AccountStatusActive},
	}, nil)

	mockOrgClient.On("DescribeAccount", mock.Anything, &organizations.DescribeAccountInput{
		AccountId: aws.String("222222222222"),
	}).Return(&organizations.DescribeAccountOutput{
		Account: &types.Account{Status: types.AccountStatusSuspended},
	}, nil)

	mockLogger.On("Logf", mock.Anything, mock.Anything).Return(nil)

	activeAccounts := filter.FilterActiveAccounts()

	assert.Len(t, activeAccounts, 1)
	assert.Equal(t, "111111111111", *activeAccounts[0].Id)

	mockOrgClient.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

func TestIsExcludedAccount(t *testing.T) {
	filter := &accountFilter{
		ExcludeAccounts: []string{"111111111111", "222222222222"},
	}

	assert.True(t, filter.isExcludedAccount("111111111111"))
	assert.True(t, filter.isExcludedAccount("222222222222"))
	assert.False(t, filter.isExcludedAccount("333333333333"))
}

func TestIsAccountActive(t *testing.T) {
	mockOrgClient := new(mocks.MockOrganizationsClient)
	mockLogger := new(mocks.MockLogger)

	filter := &accountFilter{
		OrgClient: mockOrgClient,
		Logger:    mockLogger,
	}

	mockOrgClient.On("DescribeAccount", mock.Anything, &organizations.DescribeAccountInput{
		AccountId: aws.String("111111111111"),
	}).Return(&organizations.DescribeAccountOutput{
		Account: &types.Account{Status: types.AccountStatusActive},
	}, nil)

	mockOrgClient.On("DescribeAccount", mock.Anything, &organizations.DescribeAccountInput{
		AccountId: aws.String("222222222222"),
	}).Return(&organizations.DescribeAccountOutput{
		Account: &types.Account{Status: types.AccountStatusSuspended},
	}, nil)

	mockOrgClient.On("DescribeAccount", mock.Anything, &organizations.DescribeAccountInput{
		AccountId: aws.String("333333333333"),
	}).Return(nil, errors.New("test error"))

	mockLogger.On("Logf", "Failed to describe account %s: %v", "333333333333", mock.Anything).Return(nil)

	assert.True(t, filter.isAccountActive("111111111111"))
	assert.False(t, filter.isAccountActive("222222222222"))
	assert.False(t, filter.isAccountActive("333333333333"))

	mockOrgClient.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}
