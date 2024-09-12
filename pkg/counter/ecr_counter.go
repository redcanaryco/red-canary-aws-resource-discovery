package counter

import (
	"aws-resource-discovery/pkg/interfaces"
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/service/ecr"
)

// EcrCounter is a counter for ECR repositories.
type EcrCounter struct {
	Client interfaces.ECRClient
	Result interfaces.CounterResult
}

// NewEcrCounter creates a new EcrCounter.
func NewEcrCounter(client interfaces.ECRClient) *EcrCounter {
	return &EcrCounter{
		Client: client,
		Result: interfaces.CounterResult{CounterClass: "AWS::ECR::Repository"},
	}
}

// Call performs the counting and formats the result.
func (c *EcrCounter) Call() {
	count, err := c.ecrCount()
	c.Result = c.formatResult(count, err)
	if err != nil {
		log.Printf("Error counting AWS::ECR::Repository: %v", err)
	}
}

// ecrCount counts the number of images in ECR repositories.
func (c *EcrCounter) ecrCount() (int, error) {
	input := &ecr.DescribeRepositoriesInput{}
	totalCount := 0

	for {
		result, err := c.Client.DescribeRepositories(context.TODO(), input)
		if err != nil {
			return 0, fmt.Errorf("failed to list ECR repositories: %w", err)
		}

		for _, repo := range result.Repositories {
			imageCount, err := c.countImagesInRepository(repo.RepositoryName)
			if err != nil {
				return 0, err
			}
			totalCount += imageCount
		}

		if result.NextToken == nil {
			break
		}
		input.NextToken = result.NextToken
	}

	return totalCount, nil
}

// countImagesInRepository counts the number of images in a given repository.
func (c *EcrCounter) countImagesInRepository(repositoryName *string) (int, error) {
	input := &ecr.ListImagesInput{
		RepositoryName: repositoryName,
	}
	imageCount := 0

	for {
		result, err := c.Client.ListImages(context.TODO(), input)
		if err != nil {
			return 0, fmt.Errorf("failed to list images in repository %s: %w", *repositoryName, err)
		}
		imageCount += len(result.ImageIds)
		if result.NextToken == nil {
			break
		}
		input.NextToken = result.NextToken
	}

	return imageCount, nil
}

// formatResult formats the count result and includes any error.
func (c *EcrCounter) formatResult(count int, err error) interfaces.CounterResult {
	result := interfaces.CounterResult{
		Count:        count,
		CounterClass: "AWS::ECR::Repository",
		Error:        err,
	}
	if err != nil {
		result.PermissionSuggestion = c.permissionSuggestion()
	}
	return result
}

// permissionSuggestion returns the permissions needed for counting ECR repositories.
func (c *EcrCounter) permissionSuggestion() string {
	return `
To scan ECR repositories, the provided credentials must have the following permissions:
  - ecr:DescribeRepositories
  - ecr:ListImages
`
}

// GetResult returns the counter result.
func (c *EcrCounter) GetResult() interfaces.CounterResult {
	return c.Result
}
