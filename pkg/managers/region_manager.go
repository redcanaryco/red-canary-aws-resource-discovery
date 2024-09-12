package managers

import (
	"aws-resource-discovery/pkg/interfaces"
	"context"

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
		return []string{specifiedRegion}, nil
	}

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