package scanner

import (
	"aws-resource-discovery/pkg/config"
	"aws-resource-discovery/pkg/interfaces"
	"aws-resource-discovery/pkg/utils"
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/organizations/types"
)

type ScanResult struct {
	Config      aws.Config
	Credentials aws.Credentials
	UserConfig  config.Config
	OrgAccounts []types.Account
}

type Scanner struct {
	STSClient          interfaces.STSClient
	SessionManager     interfaces.SessionManager
	RegionsManager     interfaces.RegionsManager
	CredentialsManager interfaces.CredentialsManager
	OrgDetector        interfaces.OrgDetector
	OrgClientFactory   func(cfg aws.Config) interfaces.OrganizationsClient
	Logger             interfaces.Logger
}

func NewScanner(
	stsClient interfaces.STSClient,
	sessionManager interfaces.SessionManager,
	regionsManager interfaces.RegionsManager,
	credentialsManager interfaces.CredentialsManager,
	orgDetector interfaces.OrgDetector,
	orgClientFactory func(cfg aws.Config) interfaces.OrganizationsClient,
	logger interfaces.Logger,
) *Scanner {
	return &Scanner{
		STSClient:          stsClient,
		SessionManager:     sessionManager,
		RegionsManager:     regionsManager,
		CredentialsManager: credentialsManager,
		OrgDetector:        orgDetector,
		OrgClientFactory:   orgClientFactory,
		Logger:             logger,
	}
}

func (s *Scanner) initializeScan(ctx context.Context, config config.Config) (aws.Config, aws.Credentials, []string, error) {
	cfg, initialCredentials := s.SessionManager.InitializeSessionAndCredentials(ctx, config, s.Logger)
	if cfg.Region == "" {
		log.Printf("Failed to initialize AWS session: empty region")
		return aws.Config{}, aws.Credentials{}, nil, fmt.Errorf("failed to initialize AWS session")
	}

	regions, err := s.RegionsManager.GetRegions(ctx, cfg, config.Region, s.Logger)
	if err != nil {
		log.Printf("Failed to get regions: %v", err)
		return aws.Config{}, aws.Credentials{}, nil, fmt.Errorf("failed to get regions: %w", err)
	}

	return cfg, initialCredentials, regions, nil
}

func (s *Scanner) initializeOrgScanner(cfg aws.Config, orgAccounts []types.Account, regions []string) *OrgScanner {
	orgClient := s.OrgClientFactory(cfg)
	if orgClient == nil {
		log.Printf("OrgClient is nil")
		return nil
	}
	return &OrgScanner{
		CredentialsManager: s.CredentialsManager,
		OrgAccounts:        orgAccounts,
		Logger:             s.Logger,
		Regions:            regions,
		STSClient:          s.STSClient,
		OrgClient:          orgClient,
		ScannerFactory: func(accountId, region string, credentials aws.Credentials, logger interfaces.Logger, totals *interfaces.ResourceTotals) ResourceScannerInterface {
			return &ResourceScanner{
				AccountId:   accountId,
				Region:      region,
				Credentials: credentials,
				Logger:      logger,
				Totals:      totals,
			}
		},
	}
}

func (s *Scanner) performScan(cfg aws.Config, initialCredentials aws.Credentials, regions []string, orgAccounts []types.Account, config config.Config) (ScanResult, error) {
	orgScanner := s.initializeOrgScanner(cfg, orgAccounts, regions)
	if orgScanner == nil {
		log.Printf("Failed to initialize org scanner")
		return ScanResult{}, fmt.Errorf("failed to initialize org scanner")
	}

	orgScanner.Call()
	return ScanResult{
		Config:      cfg,
		Credentials: initialCredentials,
		UserConfig:  config,
		OrgAccounts: orgAccounts,
	}, nil
}

func (s *Scanner) ScanSingleAccount(ctx context.Context, config config.Config) (ScanResult, error) {
	cfg, initialCredentials, regions, err := s.initializeScan(ctx, config)
	if err != nil {
		return ScanResult{}, err
	}
	orgAccounts := []types.Account{{Id: aws.String(config.AccountId)}}
	return s.performScan(cfg, initialCredentials, regions, orgAccounts, config)
}

func (s *Scanner) ScanOrganization(ctx context.Context, config config.Config) (ScanResult, error) {
	cfg, initialCredentials, regions, err := s.initializeScan(ctx, config)
	if err != nil {
		return ScanResult{}, err
	}

	allAccounts := s.OrgDetector.ListAccounts()
	orgClient := s.OrgClientFactory(cfg)
	if orgClient == nil {
		log.Printf("orgClient is nil")
		return ScanResult{}, fmt.Errorf("orgClient is nil")
	}

	accountFilter := utils.NewAccountFilter(allAccounts, orgClient, s.Logger, config.ExcludeAccounts)
	orgAccounts := accountFilter.FilterActiveAccounts()

	return s.performScan(cfg, initialCredentials, regions, orgAccounts, config)
}
