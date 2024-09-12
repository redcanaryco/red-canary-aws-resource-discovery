package counter

import "aws-resource-discovery/pkg/interfaces"

type RdsCounter struct {
	BaseCounter
}

func NewRdsCounter(client interfaces.CloudControlClient) *RdsCounter {
	return &RdsCounter{
		BaseCounter: BaseCounter{
			Client:   client,
			Result:   interfaces.CounterResult{CounterClass: "AWS::RDS::DBInstance"},
			TypeName: "AWS::RDS::DBInstance",
			PermissionSuggestionFunc: func() string {
				return "\nTo scan RDS instances, the provided credentials must have the following permissions:\n- rds:DescribeDBInstances\n"
			},
		},
	}
}
