package scanner

import (
	"context"
	"fmt"

	"aws-resource-discovery/pkg/interfaces"
	"aws-resource-discovery/pkg/utils"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/organizations/types"
)

type OrgScanner struct {
	CredentialsManager interfaces.CredentialsManager
	OrgAccounts        []types.Account
	Logger             interfaces.Logger
	Regions            []string
	STSClient          interfaces.STSClient
	OrgClient          interfaces.OrganizationsClient
	ScannerFactory     func(accountId, region string, credentials aws.Credentials, logger interfaces.Logger, totals *interfaces.ResourceTotals) ResourceScannerInterface
}

type ResourceScannerInterface interface {
	Call()
}

func (s *OrgScanner) Call() {
	totals := interfaces.ResourceTotals{}
	progress := newProgressReporter(len(s.OrgAccounts), len(s.Regions))

	for _, account := range s.OrgAccounts {
		for _, region := range s.Regions {
			progress.report()
			s.scanOne(&account, region, &totals)
		}
	}

	s.printSummary(totals)
}

func (s *OrgScanner) scanOne(account *types.Account, region string, totals *interfaces.ResourceTotals) {
	orgCreds, err := s.CredentialsManager.CredentialsFor(context.TODO(), *account.Id, region)
	if err != nil {
		s.Logger.Logf("Failed to get credentials for account %s in region %s: %v", *account.Id, region, err)
		return
	}

	resourceScanner := s.ScannerFactory(*account.Id, region, orgCreds, s.Logger, totals)
	resourceScanner.Call()
}

func newProgressReporter(accountsCount, regionsCount int) *progressReporter {
	return &progressReporter{
		count: 1,
		total: accountsCount * regionsCount,
	}
}

func (s *OrgScanner) printSummary(totals interfaces.ResourceTotals) {
	if len(s.OrgAccounts) == 1 {
		fmt.Println("\nScanned 1 AWS account.")
	} else {
		fmt.Printf("\nScanned %d AWS accounts.\n", len(s.OrgAccounts))
	}
	fmt.Printf("\n")
	utils.PrintTotals(totals)
}

type progressReporter struct {
	count int
	total int
}

func (p *progressReporter) report() {
	fmt.Printf("Red Canary - AWS Resource Discovery Scan Progress: %d / %d ...\r", p.count, p.total)
	p.count++
}
