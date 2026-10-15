package scanner

import (
	"bytes"
	"io"
	"os"
	"testing"

	"aws-resource-discovery/pkg/interfaces"
	"aws-resource-discovery/pkg/mocks"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/organizations/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestOrgScanner_Call(t *testing.T) {
	mockCredentialsManager := new(mocks.MockCredentialsManager)
	mockRegionsManager := new(mocks.MockRegionsManager)
	mockLogger := new(mocks.MockLogger)
	mockSTSClient := new(mocks.MockSTSClient)
	mockOrgClient := new(mocks.MockOrganizationsClient)
	mockResourceScanner := new(mocks.MockResourceScanner)

	orgAccounts := []types.Account{
		{Id: aws.String("account1")},
		{Id: aws.String("account2")},
	}

	mockCredentials := aws.Credentials{}
	// For account1: GetRegions returns [us-east-1, us-west-2]
	mockCredentialsManager.On("CredentialsFor", mock.Anything, "account1", "us-east-1").Return(mockCredentials, nil).Times(3) // Once for region discovery, twice for scanning
	mockRegionsManager.On("GetRegions", mock.Anything, mock.Anything, "", mock.Anything).Return([]string{"us-east-1", "us-west-2"}, nil).Once()
	mockCredentialsManager.On("CredentialsFor", mock.Anything, "account1", "us-west-2").Return(mockCredentials, nil).Once()

	// For account2: GetRegions returns [eu-west-1]
	mockCredentialsManager.On("CredentialsFor", mock.Anything, "account2", "us-east-1").Return(mockCredentials, nil).Once() // For region discovery
	mockRegionsManager.On("GetRegions", mock.Anything, mock.Anything, "", mock.Anything).Return([]string{"eu-west-1"}, nil).Once()
	mockCredentialsManager.On("CredentialsFor", mock.Anything, "account2", "eu-west-1").Return(mockCredentials, nil).Once()

	mockResourceScanner.On("Call").Return(nil)
	mockLogger.On("Logf", mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()

	scanner := &OrgScanner{
		CredentialsManager: mockCredentialsManager,
		RegionsManager:     mockRegionsManager,
		OrgAccounts:        orgAccounts,
		Logger:             mockLogger,
		SpecifiedRegion:    "",
		STSClient:          mockSTSClient,
		OrgClient:          mockOrgClient,
		ScannerFactory: func(accountId, region string, credentials aws.Credentials, logger interfaces.Logger, totals *interfaces.ResourceTotals, s3ScanRegion string) ResourceScannerInterface {
			return mockResourceScanner
		},
	}

	scanner.Call()

	// Total calls: 2 accounts × different region counts = 3 scans (2 for account1, 1 for account2)
	mockResourceScanner.AssertNumberOfCalls(t, "Call", 3)
	mockRegionsManager.AssertNumberOfCalls(t, "GetRegions", 2) // Once per account
}

func TestProgressReporter(t *testing.T) {
	progress := newProgressReporter(2, 2)

	expectedOutput := "Red Canary - AWS Resource Discovery Scan Progress: 1 / 4 ...\r"
	assert.Equal(t, expectedOutput, captureOutput(func() {
		progress.report()
	}))

	expectedOutput = "Red Canary - AWS Resource Discovery Scan Progress: 2 / 4 ...\r"
	assert.Equal(t, expectedOutput, captureOutput(func() {
		progress.report()
	}))
}

func TestOrgScanner_printSummary(t *testing.T) {
	mockLogger := new(mocks.MockLogger)
	mockLogger.On("Logf", mock.Anything, mock.Anything).Return(nil)

	scanner := &OrgScanner{
		Logger: mockLogger,
	}

	totals := interfaces.ResourceTotals{}
	orgAccounts := []types.Account{{Id: aws.String("account1")}}
	scanner.OrgAccounts = orgAccounts

	expectedOutput := "\nScanned 1 AWS account.\n\n"
	assert.Equal(t, expectedOutput, captureOutput(func() {
		scanner.printSummary(totals)
	}))

	orgAccounts = append(orgAccounts, types.Account{Id: aws.String("account2")})
	scanner.OrgAccounts = orgAccounts

	expectedOutput = "\nScanned 2 AWS accounts.\n\n"
	assert.Equal(t, expectedOutput, captureOutput(func() {
		scanner.printSummary(totals)
	}))
}

// Helper function to capture output
func captureOutput(f func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}
