package counter

import (
	"aws-resource-discovery/pkg/interfaces"
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/eks"
)

// EksCounter is a counter for EKS clusters.
type EksCounter struct {
	EKSClient interfaces.EKSClient
	EC2Client interfaces.EC2Client
	Result    interfaces.CounterResult
}

// NewEksCounter creates a new EksCounter.
func NewEksCounter(eksClient interfaces.EKSClient, ec2Client interfaces.EC2Client) *EksCounter {
	return &EksCounter{
		EKSClient: eksClient,
		EC2Client: ec2Client,
		Result: interfaces.CounterResult{
			CounterClass: "AWS::EKS::Cluster",
		},
	}
}

// Call performs the counting and formats the result.
func (c *EksCounter) Call() {
	count, err := c.eksCount()
	c.Result = c.formatResult(count, err)
	if err != nil {
		log.Printf("Error counting AWS::EKS::Cluster: %v", err)
	}
}

// eksCount counts the number of resources associated with EKS clusters.
func (c *EksCounter) eksCount() (int, error) {
	clusters, err := c.listClusters()
	if err != nil {
		return 0, fmt.Errorf("failed to list EKS clusters: %w", err)
	}

	totalCount := 0
	for _, clusterName := range clusters {
		count, err := c.countInstancesForCluster(clusterName)
		if err != nil {
			return 0, err
		}
		totalCount += count
	}

	return totalCount, nil
}

// listClusters lists all EKS clusters.
func (c *EksCounter) listClusters() ([]string, error) {
	clusters := []string{}
	input := &eks.ListClustersInput{}
	paginator := eks.NewListClustersPaginator(c.EKSClient, input)
	for paginator.HasMorePages() {
		output, err := paginator.NextPage(context.TODO())
		if err != nil {
			return nil, err
		}
		clusters = append(clusters, output.Clusters...)
	}
	return clusters, nil
}

// countInstancesForCluster counts the instances associated with an EKS cluster.
func (c *EksCounter) countInstancesForCluster(clusterName string) (int, error) {
	tagKey := "tag-key"
	tagValue := fmt.Sprintf("kubernetes.io/cluster/%s", clusterName)

	filters := []types.Filter{
		{
			Name:   aws.String(tagKey),
			Values: []string{tagValue},
		},
	}

	totalCount := 0
	input := &ec2.DescribeInstancesInput{
		Filters: filters,
	}

	paginator := ec2.NewDescribeInstancesPaginator(c.EC2Client, input)
	for paginator.HasMorePages() {
		output, err := paginator.NextPage(context.TODO())
		if err != nil {
			return 0, fmt.Errorf("failed to describe instances for cluster %s: %w", clusterName, err)
		}

		for _, reservation := range output.Reservations {
			totalCount += len(reservation.Instances)
		}
	}

	return totalCount, nil
}

// formatResult formats the count result and includes any error.
func (c *EksCounter) formatResult(count int, err error) interfaces.CounterResult {
	result := interfaces.CounterResult{
		Count:        count,
		CounterClass: "AWS::EKS::Cluster",
		Error:        err,
	}
	if err != nil {
		result.PermissionSuggestion = c.permissionSuggestion()
	}
	return result
}

// permissionSuggestion returns the permissions needed for counting EKS clusters.
func (c *EksCounter) permissionSuggestion() string {
	return `
To scan EKS clusters, the provided credentials must have the following permissions:
- eks:ListClusters
- ec2:DescribeInstances
`
}

// GetResult returns the counter result.
func (c *EksCounter) GetResult() interfaces.CounterResult {
	return c.Result
}
