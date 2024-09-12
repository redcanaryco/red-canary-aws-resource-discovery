package counter

import "aws-resource-discovery/pkg/interfaces"

// Ec2Counter is a counter for EC2 instances.
type Ec2Counter struct {
	BaseCounter
}

// NewEc2Counter creates a new Ec2Counter.
func NewEc2Counter(client interfaces.CloudControlClient) *Ec2Counter {
	return &Ec2Counter{
		BaseCounter: BaseCounter{
			Client:   client,
			Result:   interfaces.CounterResult{CounterClass: "AWS::EC2::Instance"},
			TypeName: "AWS::EC2::Instance",
			PermissionSuggestionFunc: func() string {
				return "\nTo scan EC2 instances, the provided credentials must have the following permissions:\n- ec2:DescribeInstances\n"
			},
		},
	}
}
