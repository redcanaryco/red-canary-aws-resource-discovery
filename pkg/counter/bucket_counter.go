package counter

import "aws-resource-discovery/pkg/interfaces"

// BucketCounter is a counter for S3 buckets.
type BucketCounter struct {
	BaseCounter
}

// NewBucketCounter creates a new BucketCounter.
func NewBucketCounter(client interfaces.CloudControlClient) *BucketCounter {
	return &BucketCounter{
		BaseCounter: BaseCounter{
			Client:   client,
			Result:   interfaces.CounterResult{CounterClass: "AWS::S3::Bucket"},
			TypeName: "AWS::S3::Bucket",
			PermissionSuggestionFunc: func() string {
				return "\nTo scan S3 buckets, the provided credentials must have the following permissions:\n- s3:ListAllMyBuckets\n- s3:GetBucketLocation\n"
			},
		},
	}
}
