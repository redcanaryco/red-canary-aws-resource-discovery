package counter

import (
	"aws-resource-discovery/pkg/interfaces"
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
)

// EcsCounter is a counter for ECS containers.
type EcsCounter struct {
	ECSClient interfaces.ECSClient
	Result    interfaces.CounterResult
}

// NewEcsCounter creates a new EcsCounter.
func NewEcsCounter(client interfaces.ECSClient) *EcsCounter {
	return &EcsCounter{
		ECSClient: client,
		Result: interfaces.CounterResult{
			CounterClass: "AWS::ECS::Cluster",
		},
	}
}

// Call performs the counting and formats the result.
func (c *EcsCounter) Call() {
	count, err := c.ecsCount()
	c.Result = c.formatResult(count, err)
	if err != nil {
		log.Printf("Error counting AWS::ECS::Cluster: %v", err)
	}
}

// ecsCount counts the number of running ECS containers.
func (c *EcsCounter) ecsCount() (int, error) {
	count := 0
	clusters, err := c.listClusters()
	if err != nil {
		return 0, fmt.Errorf("failed to list ECS clusters: %w", err)
	}

	for _, clusterName := range clusters {
		services, err := c.listServices(clusterName)
		if err != nil {
			return 0, err
		}
		for _, service := range services {
			for _, deployment := range service.Deployments {
				task, err := c.ECSClient.DescribeTaskDefinition(context.TODO(), &ecs.DescribeTaskDefinitionInput{
					TaskDefinition: deployment.TaskDefinition,
				})
				if err != nil {
					return 0, fmt.Errorf("failed to describe task definition for %s: %w", *deployment.TaskDefinition, err)
				}
				count += int(deployment.RunningCount) * len(task.TaskDefinition.ContainerDefinitions)
			}
		}
	}
	return count, nil
}

// listClusters lists all ECS clusters.
func (c *EcsCounter) listClusters() ([]string, error) {
	clusters := []string{}
	input := &ecs.ListClustersInput{}
	paginator := ecs.NewListClustersPaginator(c.ECSClient, input)
	for paginator.HasMorePages() {
		output, err := paginator.NextPage(context.TODO())
		if err != nil {
			return nil, err
		}
		clusters = append(clusters, output.ClusterArns...)
	}
	return clusters, nil
}

// listServices lists all services within a cluster.
func (c *EcsCounter) listServices(clusterName string) ([]types.Service, error) {
	services := []types.Service{}
	input := &ecs.ListServicesInput{
		Cluster: &clusterName,
	}
	paginator := ecs.NewListServicesPaginator(c.ECSClient, input)
	for paginator.HasMorePages() {
		output, err := paginator.NextPage(context.TODO())
		if err != nil {
			return nil, err
		}
		describedServices, err := c.ECSClient.DescribeServices(context.TODO(), &ecs.DescribeServicesInput{
			Cluster:  &clusterName,
			Services: output.ServiceArns,
		})
		if err != nil {
			return nil, err
		}
		services = append(services, describedServices.Services...)
	}
	return services, nil
}

// formatResult formats the count result and includes any error.
func (c *EcsCounter) formatResult(count int, err error) interfaces.CounterResult {
	result := interfaces.CounterResult{
		Count:        count,
		CounterClass: "AWS::ECS::Cluster",
		Error:        err,
	}
	if err != nil {
		result.PermissionSuggestion = c.permissionSuggestion()
	}
	return result
}

// permissionSuggestion returns the permissions needed for counting ECS containers.
func (c *EcsCounter) permissionSuggestion() string {
	return `
To scan ECS containers, the provided credentials must have the following permissions:
- ecs:ListClusters
- ecs:ListServices
- ecs:DescribeServices
- ecs:DescribeTaskDefinition
`
}

// GetResult returns the counter result.
func (c *EcsCounter) GetResult() interfaces.CounterResult {
	return c.Result
}
