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
	// ResourceScannerFactory allows tests to inject a ResourceScanner (for example a mock)
	// to avoid initializing real AWS clients during tests. If nil, the default
	// factory creates a real `ResourceScanner` which will initialize its own session.
	ResourceScannerFactory func(cfg aws.Config, accountId, region string, credentials aws.Credentials, logger interfaces.Logger, totals *interfaces.ResourceTotals) ResourceScannerInterface
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

func (s *Scanner) initializeScan(ctx context.Context, config config.Config) (aws.Config, aws.Credentials, error) {
	cfg, initialCredentials := s.SessionManager.InitializeSessionAndCredentials(ctx, config, s.Logger)
	if cfg.Region == "" {
		log.Printf("Failed to initialize AWS session: empty region")
		return aws.Config{}, aws.Credentials{}, fmt.Errorf("failed to initialize AWS session")
	}

	// Region discovery now happens per-account in OrgScanner
	return cfg, initialCredentials, nil
}

func (s *Scanner) initializeOrgScanner(cfg aws.Config, orgAccounts []types.Account, specifiedRegion string) *OrgScanner {
	orgClient := s.OrgClientFactory(cfg)
	if orgClient == nil {
		log.Printf("OrgClient is nil")
		return nil
	}
	return &OrgScanner{
		CredentialsManager: s.CredentialsManager,
		OrgAccounts:        orgAccounts,
		Logger:             s.Logger,
		RegionsManager:     s.RegionsManager,
		SessionManager:     s.SessionManager,
		SpecifiedRegion:    specifiedRegion,
		STSClient:          s.STSClient,
		OrgClient:          orgClient,
		ScannerFactory: func(accountId, region string, credentials aws.Credentials, logger interfaces.Logger, totals *interfaces.ResourceTotals, s3ScanRegion string) ResourceScannerInterface {
			if s.ResourceScannerFactory != nil {
				return s.ResourceScannerFactory(cfg, accountId, region, credentials, logger, totals)
			}
			return &ResourceScanner{
				AccountId:    accountId,
				Region:       region,
				Credentials:  credentials,
				Logger:       logger,
				Totals:       totals,
				S3ScanRegion: s3ScanRegion,
			}
		},
	}
}

func (s *Scanner) performScan(cfg aws.Config, initialCredentials aws.Credentials, orgAccounts []types.Account, config config.Config) (ScanResult, error) {
	orgScanner := s.initializeOrgScanner(cfg, orgAccounts, config.Region)
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
	cfg, initialCredentials, err := s.initializeScan(ctx, config)
	if err != nil {
		return ScanResult{}, err
	}
	orgAccounts := []types.Account{{Id: aws.String(config.AccountId)}}
	return s.performScan(cfg, initialCredentials, orgAccounts, config)
}

func (s *Scanner) ScanOrganization(ctx context.Context, config config.Config) (ScanResult, error) {
	cfg, initialCredentials, err := s.initializeScan(ctx, config)
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

	return s.performScan(cfg, initialCredentials, orgAccounts, config)
}
