package scanner

import (
	"context"
	"fmt"

	"aws-resource-discovery/pkg/interfaces"
	"aws-resource-discovery/pkg/utils"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/organizations/types"
)

type OrgScanner struct {
	CredentialsManager interfaces.CredentialsManager
	OrgAccounts        []types.Account
	Logger             interfaces.Logger
	RegionsManager     interfaces.RegionsManager
	SessionManager     interfaces.SessionManager
	SpecifiedRegion    string
	STSClient          interfaces.STSClient
	OrgClient          interfaces.OrganizationsClient
	ScannerFactory     func(accountId, region string, credentials aws.Credentials, logger interfaces.Logger, totals *interfaces.ResourceTotals, s3ScanRegion string) ResourceScannerInterface
}

type ResourceScannerInterface interface {
	Call()
}

func (s *OrgScanner) Call() {
	totals := interfaces.ResourceTotals{}
	progress := newAccountProgress(len(s.OrgAccounts))

	for _, account := range s.OrgAccounts {
		// Get regions specific to this account
		regions, err := s.getRegionsForAccount(context.TODO(), *account.Id)
		if err != nil {
			s.Logger.Logf("Failed to get regions for account %s: %v. Skipping account.", *account.Id, err)
			progress.reportSkipped(*account.Id)
			continue
		}

		if len(regions) == 0 {
			s.Logger.Logf("No enabled regions found for account %s. Skipping account.", *account.Id)
			progress.reportSkipped(*account.Id)
			continue
		}

		// S3 bucket scanning: use first available region
		s3ScanRegion := regions[0]
		progress.report(*account.Id, len(regions))

		for _, region := range regions {
			s.scanOne(&account, region, &totals, s3ScanRegion)
		}
	}

	s.printSummary(totals)
}

func (s *OrgScanner) getRegionsForAccount(ctx context.Context, accountId string) ([]string, error) {
	// Get temporary credentials to call DescribeRegions for this account
	// Use a placeholder region since we're just getting the region list
	tempCreds, err := s.CredentialsManager.CredentialsFor(ctx, accountId, "us-east-1")
	if err != nil {
		return nil, fmt.Errorf("failed to get credentials: %w", err)
	}

	// Create an AWS config with the account's credentials
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion("us-east-1"),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			tempCreds.AccessKeyID,
			tempCreds.SecretAccessKey,
			tempCreds.SessionToken,
		)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create config: %w", err)
	}

	// Call GetRegions with the account's credentials
	regions, err := s.RegionsManager.GetRegions(ctx, cfg, s.SpecifiedRegion, s.Logger)
	if err != nil {
		return nil, fmt.Errorf("failed to describe regions: %w", err)
	}

	return regions, nil
}

func (s *OrgScanner) scanOne(account *types.Account, region string, totals *interfaces.ResourceTotals, s3ScanRegion string) {
	orgCreds, err := s.CredentialsManager.CredentialsFor(context.TODO(), *account.Id, region)
	if err != nil {
		s.Logger.Logf("Failed to get credentials for account %s in region %s: %v", *account.Id, region, err)
		return
	}

	resourceScanner := s.ScannerFactory(*account.Id, region, orgCreds, s.Logger, totals, s3ScanRegion)
	resourceScanner.Call()
}

func newProgressReporter(accountsCount, regionsCount int) *progressReporter {
	return &progressReporter{
		count: 1,
		total: accountsCount * regionsCount,
	}
}

func newAccountProgress(totalAccounts int) *accountProgress {
	return &accountProgress{
		current: 0,
		total:   totalAccounts,
		skipped: 0,
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

type accountProgress struct {
	current int
	total   int
	skipped int
}

func (p *accountProgress) report(accountId string, regionCount int) {
	p.current++
	fmt.Printf("Red Canary - AWS Resource Discovery: Account %d/%d (%s) - scanning %d regions...\r", p.current, p.total, accountId, regionCount)
}

func (p *accountProgress) reportSkipped(accountId string) {
	p.skipped++
	fmt.Printf("Red Canary - AWS Resource Discovery: Skipping account %s (%d skipped so far)...\r", accountId, p.skipped)
}
