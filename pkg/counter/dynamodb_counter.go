package counter

import "aws-resource-discovery/pkg/interfaces"

// DynamoDbCounter is a counter for DynamoDB tables.
type DynamoDbCounter struct {
	BaseCounter
}

// NewDynamoDbCounter creates a new DynamoDbCounter.
func NewDynamoDbCounter(client interfaces.CloudControlClient) *DynamoDbCounter {
	return &DynamoDbCounter{
		BaseCounter: BaseCounter{
			Client:   client,
			Result:   interfaces.CounterResult{CounterClass: "AWS::DynamoDB::Table"},
			TypeName: "AWS::DynamoDB::Table",
			PermissionSuggestionFunc: func() string {
				return "\nTo scan DynamoDB tables, the provided credentials must have the following permissions:\n- dynamodb:ListTables\n- dynamodb:ListGlobalTables\n"
			},
		},
	}
}
