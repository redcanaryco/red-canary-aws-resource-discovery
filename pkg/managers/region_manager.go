package managers

import (
	"aws-resource-discovery/pkg/interfaces"
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

type RegionManager struct {
	ec2Client interfaces.EC2Client
}

func NewRegionManager(ec2Client interfaces.EC2Client) interfaces.RegionsManager {
	return &RegionManager{
		ec2Client: ec2Client,
	}
}

func (rf *RegionManager) GetRegions(ctx context.Context, cfg aws.Config, specifiedRegion string, logger interfaces.Logger) ([]string, error) {
	if specifiedRegion != "" {
		// When a specific region is requested, validate it's enabled
		allRegions, err := rf.describeAllRegions(ctx, logger)
		if err != nil {
			return nil, err
		}

		// Check if specified region is in enabled regions
		for _, region := range allRegions {
			if region == specifiedRegion {
				return []string{specifiedRegion}, nil
			}
		}

		return nil, fmt.Errorf("specified region %s is not enabled for this account", specifiedRegion)
	}

	// No region specified - return all enabled regions
	return rf.describeAllRegions(ctx, logger)
}

func (rf *RegionManager) describeAllRegions(ctx context.Context, logger interfaces.Logger) ([]string, error) {
	input := &ec2.DescribeRegionsInput{}
	resp, err := rf.ec2Client.DescribeRegions(ctx, input)
	if err != nil {
		logger.Logf("Failed to describe regions: %v", err)
		return nil, err
	}

	var regions []string
	for _, region := range resp.Regions {
		regions = append(regions, aws.ToString(region.RegionName))
	}
	return regions, nil
}