package interfaces

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/cloudtrail"
)

type CloudTrailClient interface {
	DescribeTrails(ctx context.Context, input *cloudtrail.DescribeTrailsInput, opts ...func(*cloudtrail.Options)) (*cloudtrail.DescribeTrailsOutput, error)
}
