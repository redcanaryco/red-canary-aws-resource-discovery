package counter

import "aws-resource-discovery/pkg/interfaces"

// EbsCounter is a counter for EBS volumes.
type EbsCounter struct {
	BaseCounter
}

// NewEbsCounter creates a new EbsCounter.
func NewEbsCounter(client interfaces.CloudControlClient) *EbsCounter {
	return &EbsCounter{
		BaseCounter: BaseCounter{
			Client:   client,
			Result:   interfaces.CounterResult{CounterClass: "AWS::EC2::Volume"},
			TypeName: "AWS::EC2::Volume",
			PermissionSuggestionFunc: func() string {
				return "\nTo scan EBS volumes, the provided credentials must have the following permissions:\n- ec2:DescribeVolumes\n"
			},
		},
	}
}
