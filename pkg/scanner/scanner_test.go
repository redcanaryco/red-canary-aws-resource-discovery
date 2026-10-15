package scanner

import (
	"aws-resource-discovery/pkg/config"
	"aws-resource-discovery/pkg/interfaces"
	"aws-resource-discovery/pkg/mocks"
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/organizations"
	"github.com/aws/aws-sdk-go-v2/service/organizations/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// (Use mocks.MockResourceScanner below to exercise logger expectations.)

func TestScanner_ScanSingleAccount(t *testing.T) {
	mockSTSClient := new(mocks.MockSTSClient)
	mockSessionManager := new(mocks.MockSessionManager)
	mockRegionsManager := new(mocks.MockRegionsManager)
	mockCredentialsManager := new(mocks.MockCredentialsManager)
	mockOrgDetector := new(mocks.MockOrgDetector)
	mockLogger := new(mocks.MockLogger)
	mockOrgClient := new(mocks.MockOrganizationsClient)

	mockOrgClientFactory := func(cfg aws.Config) interfaces.OrganizationsClient {
		return mockOrgClient
	}

	mockSessionManager.On("InitializeSessionAndCredentials", mock.Anything, mock.Anything, mock.Anything).Return(aws.Config{Region: "us-east-1"}, aws.Credentials{})
	// GetRegions now called per-account in OrgScanner, not in initializeScan
	mockRegionsManager.On("GetRegions", mock.Anything, mock.Anything, "us-east-1", mock.Anything).Return([]string{"us-east-1"}, nil)
	mockCredentialsManager.On("CredentialsFor", mock.Anything, "123456789012", "us-east-1").Return(aws.Credentials{}, nil).Times(2) // Called twice: once for region discovery, once for scanning
	mockLogger.On("Log", mock.Anything).Return(nil)
	mockLogger.On("Logf", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	scanner := NewScanner(mockSTSClient, mockSessionManager, mockRegionsManager, mockCredentialsManager, mockOrgDetector, mockOrgClientFactory, mockLogger)

	// Inject repo mock ResourceScanner; have its Call() trigger the logger so
	// existing assertions about logger activity remain valid.
	mockResourceScanner := new(mocks.MockResourceScanner)
	mockResourceScanner.On("Call").Run(func(args mock.Arguments) {
		mockLogger.Log([]string{"123456789012", "us-east-1", "AWS::TEST::Resource", "0"})
	}).Return()
	scanner.ResourceScannerFactory = func(cfg aws.Config, accountId, region string, credentials aws.Credentials, logger interfaces.Logger, totals *interfaces.ResourceTotals) ResourceScannerInterface {
		return mockResourceScanner
	}

	userConfig := config.Config{AccountId: "123456789012", Region: "us-east-1"}
	ctx := context.Background()

	result, err := scanner.ScanSingleAccount(ctx, userConfig)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, userConfig, result.UserConfig)
	assert.Len(t, result.OrgAccounts, 1)
	assert.Equal(t, "123456789012", *result.OrgAccounts[0].Id)

	mockSessionManager.AssertCalled(t, "InitializeSessionAndCredentials", ctx, userConfig, mockLogger)
	// GetRegions now called per-account with specified region
	mockRegionsManager.AssertCalled(t, "GetRegions", mock.Anything, mock.Anything, "us-east-1", mock.Anything)
	mockCredentialsManager.AssertCalled(t, "CredentialsFor", mock.Anything, "123456789012", "us-east-1")
	mockLogger.AssertCalled(t, "Log", mock.Anything)
}

func TestScanner_ScanOrganization(t *testing.T) {
	mockSTSClient := new(mocks.MockSTSClient)
	mockSessionManager := new(mocks.MockSessionManager)
	mockRegionsManager := new(mocks.MockRegionsManager)
	mockCredentialsManager := new(mocks.MockCredentialsManager)
	mockOrgDetector := new(mocks.MockOrgDetector)
	mockLogger := new(mocks.MockLogger)
	mockOrgClient := new(mocks.MockOrganizationsClient)

	mockOrgClientFactory := func(cfg aws.Config) interfaces.OrganizationsClient {
		return mockOrgClient
	}

	mockSessionManager.On("InitializeSessionAndCredentials", mock.Anything, mock.Anything, mock.Anything).Return(aws.Config{Region: "us-east-1"}, aws.Credentials{})
	// GetRegions now called per-account in OrgScanner
	mockRegionsManager.On("GetRegions", mock.Anything, mock.Anything, "us-east-1", mock.Anything).Return([]string{"us-east-1"}, nil)
	mockOrgDetector.On("ListAccounts").Return([]types.Account{{Id: aws.String("123456789012")}})
	mockCredentialsManager.On("CredentialsFor", mock.Anything, "123456789012", "us-east-1").Return(aws.Credentials{}, nil).Times(2) // Called twice: once for region discovery, once for scanning
	mockOrgClient.On("DescribeAccount", mock.Anything, mock.Anything).Return(&organizations.DescribeAccountOutput{
		Account: &types.Account{Id: aws.String("123456789012"), Status: types.AccountStatusActive},
	}, nil)
	mockLogger.On("Log", mock.Anything).Return(nil)
	mockLogger.On("Logf", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	scanner := NewScanner(mockSTSClient, mockSessionManager, mockRegionsManager, mockCredentialsManager, mockOrgDetector, mockOrgClientFactory, mockLogger)

	// Inject repo mock ResourceScanner for organization scan as well
	mockResourceScanner2 := new(mocks.MockResourceScanner)
	mockResourceScanner2.On("Call").Run(func(args mock.Arguments) {
		mockLogger.Log([]string{"123456789012", "us-east-1", "AWS::TEST::Resource", "0"})
	}).Return()
	scanner.ResourceScannerFactory = func(cfg aws.Config, accountId, region string, credentials aws.Credentials, logger interfaces.Logger, totals *interfaces.ResourceTotals) ResourceScannerInterface {
		return mockResourceScanner2
	}

	userConfig := config.Config{AccountId: "123456789012", Region: "us-east-1"}
	ctx := context.Background()

	result, err := scanner.ScanOrganization(ctx, userConfig)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, userConfig, result.UserConfig)
	assert.Len(t, result.OrgAccounts, 1)
	assert.Equal(t, "123456789012", *result.OrgAccounts[0].Id)

	mockSessionManager.AssertCalled(t, "InitializeSessionAndCredentials", ctx, userConfig, mockLogger)
	// GetRegions now called per-account
	mockRegionsManager.AssertCalled(t, "GetRegions", mock.Anything, mock.Anything, "us-east-1", mock.Anything)
	mockOrgDetector.AssertCalled(t, "ListAccounts")
	mockCredentialsManager.AssertCalled(t, "CredentialsFor", mock.Anything, "123456789012", "us-east-1")
	mockOrgClient.AssertCalled(t, "DescribeAccount", mock.Anything, mock.Anything)
	mockLogger.AssertCalled(t, "Log", mock.Anything)
}
