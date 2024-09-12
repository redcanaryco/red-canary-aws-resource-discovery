package interfaces

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
)

type RegionsManager interface {
	GetRegions(ctx context.Context, cfg aws.Config, specifiedRegion string, logger Logger) ([]string, error)
}
