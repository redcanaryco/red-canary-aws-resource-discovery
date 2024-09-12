package counter

import (
	"aws-resource-discovery/pkg/interfaces"
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudcontrol"
)

// BaseCounter provides common functionality for all counters.
type BaseCounter struct {
	Client                   interfaces.CloudControlClient
	Result                   interfaces.CounterResult
	TypeName                 string
	PermissionSuggestionFunc func() string
}

// Call performs the counting and formats the result.
func (b *BaseCounter) Call() {
	count := b.paginatedCount()
	b.Result = b.formatResult(count, b.Result.Error)
	if b.Result.Error != nil {
		log.Printf("Error counting %s: %v", b.TypeName, b.Result.Error)
	}
}

// paginatedCount handles paginated API requests to count resources.
func (b *BaseCounter) paginatedCount() int {
	input := &cloudcontrol.ListResourcesInput{
		TypeName: aws.String(b.TypeName),
	}
	count := 0
	for {
		result, err := b.Client.ListResources(context.TODO(), input)
		if err != nil {
			b.Result.Error = err
			return 0
		}
		count += len(result.ResourceDescriptions)
		if result.NextToken == nil {
			break
		}
		input.NextToken = result.NextToken
	}
	return count
}

// formatResult formats the count result and includes any error.
func (b *BaseCounter) formatResult(count int, err error) interfaces.CounterResult {
	result := interfaces.CounterResult{
		Count:        count,
		CounterClass: b.TypeName,
		Error:        err,
	}
	if err != nil && b.PermissionSuggestionFunc != nil {
		result.PermissionSuggestion = b.PermissionSuggestionFunc()
	}
	return result
}

// GetResult returns the counter result.
func (b *BaseCounter) GetResult() interfaces.CounterResult {
	return b.Result
}
