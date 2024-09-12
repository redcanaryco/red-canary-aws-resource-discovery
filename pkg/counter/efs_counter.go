package counter

import "aws-resource-discovery/pkg/interfaces"

// EfsCounter is a counter for EFS file systems.
type EfsCounter struct {
	BaseCounter
}

// NewEfsCounter creates a new EfsCounter.
func NewEfsCounter(client interfaces.CloudControlClient) *EfsCounter {
	return &EfsCounter{
		BaseCounter: BaseCounter{
			Client:   client,
			Result:   interfaces.CounterResult{CounterClass: "AWS::EFS::FileSystem"},
			TypeName: "AWS::EFS::FileSystem",
			PermissionSuggestionFunc: func() string {
				return "\nTo scan EFS file systems, the provided credentials must have the following permissions:\n- elasticfilesystem:DescribeFileSystems\n"
			},
		},
	}
}
