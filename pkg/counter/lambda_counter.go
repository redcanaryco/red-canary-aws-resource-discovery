package counter

import "aws-resource-discovery/pkg/interfaces"

// LambdaCounter is a counter for Lambda functions.
type LambdaCounter struct {
	BaseCounter
}

// NewLambdaCounter creates a new LambdaCounter.
func NewLambdaCounter(client interfaces.CloudControlClient) *LambdaCounter {
	return &LambdaCounter{
		BaseCounter: BaseCounter{
			Client:   client,
			Result:   interfaces.CounterResult{CounterClass: "AWS::Lambda::Function"},
			TypeName: "AWS::Lambda::Function",
			PermissionSuggestionFunc: func() string {
				return "\nTo scan Lambda functions, the provided credentials must have the following permissions:\n- lambda:ListFunctions\n"
			},
		},
	}
}
