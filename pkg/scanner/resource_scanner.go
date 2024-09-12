package scanner

import (
	"context"
	"strconv"

	"aws-resource-discovery/pkg/counter"
	"aws-resource-discovery/pkg/interfaces"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/cloudcontrol"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
	"github.com/aws/aws-sdk-go-v2/service/ecrpublic"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go-v2/service/eks"
)

type ResourceScanner struct {
	Session     aws.Config
	Credentials aws.Credentials
	AccountId   string
	Region      string
	Logger      interfaces.Logger
	Totals      *interfaces.ResourceTotals
}

const GLOBAL_SCAN_REGION = "us-east-1"

func (s *ResourceScanner) Call() {
	s.scanResources()
}

func (s *ResourceScanner) scanResources() {
	ctx := context.Background()
	s.Session = s.createSession(ctx)
	client := cloudcontrol.NewFromConfig(s.Session)
	eksClient := eks.NewFromConfig(s.Session)
	ec2Client := ec2.NewFromConfig(s.Session)
	ecsClient := ecs.NewFromConfig(s.Session)
	ecrClient := ecr.NewFromConfig(s.Session)
	var counters []interfaces.Counter

	// Only add the BucketCounter if the region is us-east-1
	// ECR Public is only available in us-east-1
	if s.Region == GLOBAL_SCAN_REGION {
		counters = append(counters, counter.NewBucketCounter(client))

		// Saves on API calls if we don't need to scan ECR Public
		client_ecrpublic := ecrpublic.NewFromConfig(s.Session)
		counters = append(counters, counter.NewEcrPublicCounter(client_ecrpublic))
	}

	counters = append(counters,
		counter.NewEc2Counter(client),
		counter.NewDynamoDbCounter(client),
		counter.NewEbsCounter(client),
		counter.NewEcrCounter(ecrClient),
		counter.NewEcsCounter(ecsClient),
		counter.NewEfsCounter(client),
		counter.NewEksCounter(eksClient, ec2Client),
		counter.NewLambdaCounter(client),
		counter.NewRdsCounter(client))

	results := make(chan interfaces.CounterResult, len(counters))

	for _, cnt := range counters {
		go func(cnt interfaces.Counter) {
			cnt.Call()
			results <- cnt.GetResult()
		}(cnt)
	}

	resourceResults := make(map[string]int)

	for range counters {
		result := <-results
		resourceResults[result.CounterClass] = result.Count
		s.updateTotals(result.CounterClass, result.Count)
	}

	for resourceType, resourceCount := range resourceResults {
		record := []string{s.AccountId, s.Region, resourceType, strconv.Itoa(resourceCount)}
		s.Logger.Log(record)
	}
}

func (s *ResourceScanner) createSession(ctx context.Context) aws.Config {
	if s.Session.Region != "" {
		return s.Session
	}

	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(s.Region))
	if err != nil {
		s.Logger.Logf("Failed to initialize AWS session: %v", err)
	}

	if s.Credentials.HasKeys() {
		creds := aws.NewCredentialsCache(credentials.NewStaticCredentialsProvider(
			s.Credentials.AccessKeyID,
			s.Credentials.SecretAccessKey,
			s.Credentials.SessionToken,
		))
		cfg.Credentials = creds
	}
	return cfg
}

func (s *ResourceScanner) updateTotals(resourceType string, count int) {
	switch resourceType {
	case "AWS::S3::Bucket":
		s.Totals.Buckets += count
	case "AWS::EKS::Cluster":
		s.Totals.ContainerHosts += count
	case "AWS::DynamoDB::Table", "AWS::RDS::DBInstance":
		s.Totals.Databases += count
	case "AWS::EFS::FileSystem", "AWS::EC2::Volume":
		s.Totals.NonOsDisks += count
	case "AWS::ECS::Cluster":
		s.Totals.ServerlessContainers += count
	case "AWS::Lambda::Function":
		s.Totals.ServerlessFunctions += count
	case "AWS::EC2::Instance":
		s.Totals.VirtualMachines += count
	case "AWS::ECR::Repository", "AWS::ECR::PublicRepository":
		s.Totals.ContainerRegistryImages += count
	}
}
