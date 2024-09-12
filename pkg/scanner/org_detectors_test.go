package scanner

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

func TestOrgDetector_ListAccounts(t *testing.T) {
	mockClient := new(mocks.MockOrganizationsClient)
	mockLogger := new(mocks.MockLogger)
	detector := NewOrgDetector(mockClient, mockLogger)

	// Test successful listing
	mockClient.On("ListAccounts", mock.Anything, mock.Anything).Return(&organizations.ListAccountsOutput{
		Accounts: []types.Account{
			{Id: aws.String("123456789012")},
			{Id: aws.String("210987654321")},
		},
	}, nil).Once()

	accounts := detector.ListAccounts()
	assert.Len(t, accounts, 2)
	assert.Equal(t, "123456789012", *accounts[0].Id)
	assert.Equal(t, "210987654321", *accounts[1].Id)

	// Test AWSOrganizationsNotInUseException
	mockClient.On("ListAccounts", mock.Anything, mock.Anything).Return(nil, &types.AWSOrganizationsNotInUseException{}).Once()
	accounts = detector.ListAccounts()
	assert.Nil(t, accounts)

	// Test AccessDeniedException
	mockClient.On("ListAccounts", mock.Anything, mock.Anything).Return(nil, &types.AccessDeniedException{}).Once()
	mockLogger.On("Logf", mock.AnythingOfType("string"), mock.Anything).Return(nil).Once()
	accounts = detector.ListAccounts()
	assert.Nil(t, accounts)

	// Test other error
	mockClient.On("ListAccounts", mock.Anything, mock.Anything).Return(nil, errors.New("unknown error")).Once()
	mockLogger.On("Logf", mock.AnythingOfType("string"), mock.Anything).Return(nil).Once()
	accounts = detector.ListAccounts()
	assert.Nil(t, accounts)

	mockClient.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}
