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
	mockLogger := new(mocks.MockLogger)
	mockSTSClient := new(mocks.MockSTSClient)
	mockOrgClient := new(mocks.MockOrganizationsClient)
	mockResourceScanner := new(mocks.MockResourceScanner)

	orgAccounts := []types.Account{
		{Id: aws.String("account1")},
		{Id: aws.String("account2")},
	}
	regions := []string{"us-east-1", "us-west-2"}

	mockCredentials := aws.Credentials{}
	mockCredentialsManager.On("CredentialsFor", mock.Anything, "account1", "us-east-1").Return(mockCredentials, nil)
	mockCredentialsManager.On("CredentialsFor", mock.Anything, "account1", "us-west-2").Return(mockCredentials, nil)
	mockCredentialsManager.On("CredentialsFor", mock.Anything, "account2", "us-east-1").Return(mockCredentials, nil)
	mockCredentialsManager.On("CredentialsFor", mock.Anything, "account2", "us-west-2").Return(mockCredentials, nil)

	mockResourceScanner.On("Call").Return(nil)

	scanner := &OrgScanner{
		CredentialsManager: mockCredentialsManager,
		OrgAccounts:        orgAccounts,
		Logger:             mockLogger,
		Regions:            regions,
		STSClient:          mockSTSClient,
		OrgClient:          mockOrgClient,
		ScannerFactory: func(accountId, region string, credentials aws.Credentials, logger interfaces.Logger, totals *interfaces.ResourceTotals) ResourceScannerInterface {
			return mockResourceScanner
		},
	}

	scanner.Call()

	mockCredentialsManager.AssertNumberOfCalls(t, "CredentialsFor", 4)
	mockResourceScanner.AssertNumberOfCalls(t, "Call", 4)
	mockLogger.AssertNotCalled(t, "Logf")
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
