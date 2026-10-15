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
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go-v2/service/eks"
)

type ResourceScanner struct {
	Session      aws.Config
	Credentials  aws.Credentials
	AccountId    string
	Region       string
	Logger       interfaces.Logger
	Totals       *interfaces.ResourceTotals
	S3ScanRegion string // Region to use for S3 bucket scanning (empty means don't scan S3)
}

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
	var counters []interfaces.Counter

	// S3 buckets are global resources - count them once per account
	// Only include S3 counter if this is the designated S3 scan region
	if s.S3ScanRegion != "" && s.Region == s.S3ScanRegion {
		counters = append(counters, counter.NewBucketCounter(client))
	}

	counters = append(counters,
		counter.NewEc2Counter(client),
		counter.NewDynamoDbCounter(client),
		counter.NewEbsCounter(client),
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
	}
}
