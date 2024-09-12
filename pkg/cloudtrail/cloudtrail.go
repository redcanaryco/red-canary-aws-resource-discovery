package cloudtrail

import (
	"aws-resource-discovery/pkg/interfaces"
	"aws-resource-discovery/pkg/scanner"
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudtrail"
	"github.com/aws/aws-sdk-go-v2/service/cloudtrail/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/fatih/color"
	"github.com/rodaine/table"
)

type TrailInfo struct {
	AccountID           string
	TrailArn            string
	IsOrganizationTrail bool
	S3BucketArn         string
	SNSTopicArn         string
}

type CloudTrailChecker struct {
	CTClient  interfaces.CloudTrailClient
	S3Client  interfaces.S3Client
	AWSConfig aws.Config
}

func (c *CloudTrailChecker) describeTrails(ctx context.Context) (*cloudtrail.DescribeTrailsOutput, error) {
	input := &cloudtrail.DescribeTrailsInput{}
	return c.CTClient.DescribeTrails(ctx, input)
}

func (c *CloudTrailChecker) CheckCloudTrail(ctx context.Context, scanResult scanner.ScanResult) ([]TrailInfo, error) {
	trails, err := c.describeTrails(ctx)
	if err != nil {
		return nil, fmt.Errorf("error checking CloudTrail for account %s: %w", scanResult.UserConfig.AccountId, err)
	}

	return c.extractTrailInfos(ctx, trails, scanResult.UserConfig.AccountId)
}

func (c *CloudTrailChecker) extractTrailInfos(ctx context.Context, result *cloudtrail.DescribeTrailsOutput, accountId string) ([]TrailInfo, error) {
	var trailInfos []TrailInfo
	for _, trail := range result.TrailList {
		trailInfo, err := c.createTrailInfo(ctx, accountId, trail)
		if err != nil {
			return nil, fmt.Errorf("error creating trail info for account %s: %w", accountId, err)
		}
		trailInfos = append(trailInfos, trailInfo)
	}
	return trailInfos, nil
}

func (c *CloudTrailChecker) createTrailInfo(ctx context.Context, accountId string, trail types.Trail) (TrailInfo, error) {
	trailInfo := TrailInfo{
		AccountID:           accountId,
		TrailArn:            aws.ToString(trail.TrailARN),
		IsOrganizationTrail: aws.ToBool(trail.IsOrganizationTrail),
	}

	if trail.S3BucketName != nil {
		trailInfo.S3BucketArn = fmt.Sprintf("arn:aws:s3:::%s", aws.ToString(trail.S3BucketName))
		bucketRegion, err := c.getBucketRegion(ctx, aws.ToString(trail.S3BucketName))
		if err != nil {
			return trailInfo, fmt.Errorf("failed to get bucket region for bucket %s: %w", aws.ToString(trail.S3BucketName), err)
		}

		if bucketRegion != "" {
			snsArn, err := c.getSnsTopicArn(ctx, aws.ToString(trail.S3BucketName), bucketRegion)
			if err != nil {
				return trailInfo, fmt.Errorf("failed to get SNS topic ARN for bucket %s: %w", aws.ToString(trail.S3BucketName), err)
			}
			trailInfo.SNSTopicArn = snsArn
		}
	}

	return trailInfo, nil
}

func (c *CloudTrailChecker) getSnsTopicArn(ctx context.Context, bucketName, bucketRegion string) (string, error) {
	s3Client := s3.NewFromConfig(c.AWSConfig, func(o *s3.Options) {
		o.Region = bucketRegion
	})

	input := &s3.GetBucketNotificationConfigurationInput{
		Bucket: aws.String(bucketName),
	}
	notifConfig, err := s3Client.GetBucketNotificationConfiguration(ctx, input)
	if err != nil {
		return "", fmt.Errorf("failed to get bucket notification configuration for bucket %s: %w", bucketName, err)
	}

	for _, snsConfig := range notifConfig.TopicConfigurations {
		if snsConfig.TopicArn != nil {
			return aws.ToString(snsConfig.TopicArn), nil
		}
	}

	return "", nil
}

func (c *CloudTrailChecker) getBucketRegion(ctx context.Context, bucketName string) (string, error) {
	input := &s3.GetBucketLocationInput{
		Bucket: aws.String(bucketName),
	}
	locationOutput, err := c.S3Client.GetBucketLocation(ctx, input)
	if err != nil {
		return "", fmt.Errorf("failed to get bucket location for bucket %s: %w", bucketName, err)
	}

	if locationOutput.LocationConstraint == "" {
		return "us-east-1", nil
	}

	return string(locationOutput.LocationConstraint), nil
}

func PrintTable(ctx context.Context, trailInfos []TrailInfo) {
	headerFmt := color.New(color.FgGreen, color.Underline).SprintfFunc()
	columnFmt := color.New(color.FgYellow).SprintfFunc()
	tbl := table.New("ResourceType", "ARN").WithPadding(3)
	tbl.WithHeaderFormatter(headerFmt).WithFirstColumnFormatter(columnFmt)
	for _, info := range trailInfos {
		if info.TrailArn != "" {
			tbl.AddRow("Trail", info.TrailArn)
		}
		if info.S3BucketArn != "" {
			tbl.AddRow("S3", info.S3BucketArn)
		}
		if info.SNSTopicArn != "" {
			tbl.AddRow("SNS", info.SNSTopicArn)
		}
	}

	tbl.Print()
}
