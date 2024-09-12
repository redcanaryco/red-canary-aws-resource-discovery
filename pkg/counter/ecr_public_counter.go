package counter

import (
	"aws-resource-discovery/pkg/interfaces"
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/service/ecrpublic"
)

// EcrPublicCounter is a counter for ECR public repositories.
type EcrPublicCounter struct {
	Client interfaces.ECRPublicClient
	Result interfaces.CounterResult
}

// NewEcrPublicCounter creates a new EcrPublicCounter.
func NewEcrPublicCounter(client interfaces.ECRPublicClient) *EcrPublicCounter {
	return &EcrPublicCounter{
		Client: client,
		Result: interfaces.CounterResult{CounterClass: "AWS::ECR::PublicRepository"},
	}
}

// Call performs the counting and formats the result.
func (c *EcrPublicCounter) Call() {
	count, err := c.ecrPublicCount()
	c.Result = c.formatResult(count, err)
	if err != nil {
		log.Printf("Error counting AWS::ECR::PublicRepository: %v", err)
	}
}

// ecrPublicCount counts the number of images in public ECR repositories.
func (c *EcrPublicCounter) ecrPublicCount() (int, error) {
	input := &ecrpublic.DescribeRepositoriesInput{}
	totalCount := 0

	for {
		result, err := c.Client.DescribeRepositories(context.TODO(), input)
		if err != nil {
			return 0, fmt.Errorf("failed to list public ECR repositories: %w", err)
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
func (c *EcrPublicCounter) countImagesInRepository(repositoryName *string) (int, error) {
	input := &ecrpublic.DescribeImagesInput{
		RepositoryName: repositoryName,
	}
	imageCount := 0

	for {
		result, err := c.Client.DescribeImages(context.TODO(), input)
		if err != nil {
			return 0, fmt.Errorf("failed to describe images in repository %s: %w", *repositoryName, err)
		}
		imageCount += len(result.ImageDetails)
		if result.NextToken == nil {
			break
		}
		input.NextToken = result.NextToken
	}

	return imageCount, nil
}

// formatResult formats the count result and includes any error.
func (c *EcrPublicCounter) formatResult(count int, err error) interfaces.CounterResult {
	result := interfaces.CounterResult{
		Count:        count,
		CounterClass: "AWS::ECR::PublicRepository",
		Error:        err,
	}
	if err != nil {
		result.PermissionSuggestion = c.permissionSuggestion()
	}
	return result
}

// permissionSuggestion returns the permissions needed for counting ECR public repositories.
func (c *EcrPublicCounter) permissionSuggestion() string {
	return `
To scan public ECR repositories, the provided credentials must have the following permissions:
  - ecr-public:DescribeRepositories
  - ecr-public:DescribeImages
`
}

// GetResult returns the counter result.
func (c *EcrPublicCounter) GetResult() interfaces.CounterResult {
	return c.Result
}
