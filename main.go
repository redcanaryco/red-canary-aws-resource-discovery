package main

import (
	"aws-resource-discovery/pkg/cloudtrail"
	"aws-resource-discovery/pkg/config"
	"aws-resource-discovery/pkg/interfaces"
	"aws-resource-discovery/pkg/logger"
	"aws-resource-discovery/pkg/managers"
	"aws-resource-discovery/pkg/scanner"
	"context"
	"flag"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	aws_config "github.com/aws/aws-sdk-go-v2/config"
	aws_trail "github.com/aws/aws-sdk-go-v2/service/cloudtrail"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/organizations"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

func main() {
	ctx := context.Background()
	userConfig := parseFlags()

	// Setup CSV Logger
	csvLogger, err := logger.NewCSVLogger("aws-resource-discovery.csv")
	if err != nil {
		log.Fatalf("Failed to initialize CSV logger: %v", err)
	}
	defer csvLogger.Close()

	// Load the AWS SDK configuration
	cfg, err := aws_config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Fatalf("Failed to load AWS configuration: %v", err)
	}

	// Create STS Client
	stsClient := sts.NewFromConfig(cfg)
	credsManager := managers.NewCredentialsManager(userConfig.RoleName, stsClient)
	sessionManager := managers.NewSessionManager(stsClient)

	// Create EC2 Client
	ec2Client := ec2.NewFromConfig(cfg)

	// Create RegionsManager with EC2 Client
	regionsManager := managers.NewRegionManager(ec2Client)

	// Create Organizations Client
	orgClient := organizations.NewFromConfig(cfg)
	orgDetector := scanner.NewOrgDetector(orgClient, csvLogger)

	// Create Scanner service with factory function for ResourceScanner
	scanService := scanner.NewScanner(
		stsClient,
		sessionManager,
		regionsManager,
		credsManager,
		orgDetector,
		func(cfg aws.Config) interfaces.OrganizationsClient {
			return organizations.NewFromConfig(cfg)
		},
		csvLogger,
	)

	startTime := time.Now()

	// Determine the flow based on the parsed configuration
	var scanResult scanner.ScanResult
	if userConfig.AccountId != "" {
		fmt.Println("Single account scan selected.")
		scanResult, err = scanService.ScanSingleAccount(ctx, userConfig)
	} else {
		fmt.Println("Organization scan selected.")
		scanResult, err = scanService.ScanOrganization(ctx, userConfig)
	}

	if err != nil {
		log.Fatalf("Failed to perform scan: %v", err)
	}

	endTime := time.Now()
	duration := endTime.Sub(startTime)
	minutes := int(duration.Minutes())
	seconds := int(duration.Seconds()) % 60

	if minutes > 0 {
		fmt.Printf("\nScan completed in %d minutes %d seconds.\n", minutes, seconds)
	} else {
		fmt.Printf("\nScan completed in %d seconds.\n", seconds)
	}

	// Perform CloudTrail check based on the scan result
	if userConfig.Trail {
		ctClient := aws_trail.NewFromConfig(cfg)
		s3Client := s3.NewFromConfig(cfg)

		ctChecker := &cloudtrail.CloudTrailChecker{
			CTClient:  ctClient,
			S3Client:  s3Client,
			AWSConfig: cfg,
		}

		trailInfos, err := ctChecker.CheckCloudTrail(ctx, scanResult)
		if err != nil {
			log.Fatalf("Failed to check CloudTrail: %v", err)
		}
		cloudtrail.PrintTable(ctx, trailInfos)
	}
}

func parseFlags() config.Config {
	var config config.Config
	var excludeAccounts string

	flag.StringVar(&config.RoleArn, "AWS_ROLE_ARN", "", "AWS Role ARN to assume")
	flag.StringVar(&config.AccountId, "AWS_ACCOUNT_ID", "", "AWS Account ID")
	flag.StringVar(&config.Region, "AWS_REGION", "", "AWS Region")
	flag.StringVar(&config.RoleName, "AWS_ROLE_NAME", "", "AWS Role Name")
	flag.BoolVar(&config.Trail, "AWS_TRAIL", false, "Set to true to print CloudTrail information")
	flag.StringVar(&excludeAccounts, "EXCLUDE", "", "Comma-separated list of AWS account numbers to exclude")

	// Parse flags
	flag.Parse()

	// Split the EXCLUDE_ACCOUNT flag value into a slice of strings
	if excludeAccounts != "" {
		config.ExcludeAccounts = strings.Split(excludeAccounts, ",")
	}

	return config
}
