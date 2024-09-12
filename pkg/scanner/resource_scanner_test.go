package scanner

import (
	"aws-resource-discovery/pkg/interfaces"
	"aws-resource-discovery/pkg/mocks"
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestResourceScanner_createSession(t *testing.T) {
	mockLogger := new(mocks.MockLogger)

	scanner := &ResourceScanner{
		Region: "us-west-2",
		Logger: mockLogger,
	}

	mockLogger.On("Logf", mock.Anything, mock.Anything).Return()

	ctx := context.Background()
	session := scanner.createSession(ctx)

	assert.Equal(t, "us-west-2", session.Region)
	mockLogger.AssertNotCalled(t, "Logf")
}

func TestResourceScanner_createSessionWithCredentials(t *testing.T) {
	mockLogger := new(mocks.MockLogger)

	scanner := &ResourceScanner{
		Region: "us-west-2",
		Logger: mockLogger,
		Credentials: aws.Credentials{
			AccessKeyID:     "AKIAIOSFODNN7EXAMPLE",
			SecretAccessKey: "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
		},
	}

	mockLogger.On("Logf", mock.Anything, mock.Anything).Return()

	ctx := context.Background()
	session := scanner.createSession(ctx)

	assert.Equal(t, "us-west-2", session.Region)
	assert.NotNil(t, session.Credentials)
	mockLogger.AssertNotCalled(t, "Logf")
}

func TestResourceScanner_updateTotals(t *testing.T) {
	totals := &interfaces.ResourceTotals{}
	scanner := &ResourceScanner{Totals: totals}

	testCases := []struct {
		resourceType  string
		count         int
		expectedField string
		expectedValue int
	}{
		{"AWS::S3::Bucket", 5, "Buckets", 5},
		{"AWS::EKS::Cluster", 2, "ContainerHosts", 2},
		{"AWS::DynamoDB::Table", 3, "Databases", 3},
		{"AWS::RDS::DBInstance", 2, "Databases", 5},
		{"AWS::EFS::FileSystem", 1, "NonOsDisks", 1},
		{"AWS::EC2::Volume", 4, "NonOsDisks", 5},
		{"AWS::ECS::Cluster", 1, "ServerlessContainers", 1},
		{"AWS::Lambda::Function", 10, "ServerlessFunctions", 10},
		{"AWS::EC2::Instance", 3, "VirtualMachines", 3},
		{"AWS::ECR::Repository", 7, "ContainerRegistryImages", 7},
		{"AWS::ECR::PublicRepository", 2, "ContainerRegistryImages", 9},
	}

	for _, tc := range testCases {
		t.Run(tc.resourceType, func(t *testing.T) {
			scanner.updateTotals(tc.resourceType, tc.count)

			switch tc.expectedField {
			case "Buckets":
				assert.Equal(t, tc.expectedValue, totals.Buckets)
			case "ContainerHosts":
				assert.Equal(t, tc.expectedValue, totals.ContainerHosts)
			case "Databases":
				assert.Equal(t, tc.expectedValue, totals.Databases)
			case "NonOsDisks":
				assert.Equal(t, tc.expectedValue, totals.NonOsDisks)
			case "ServerlessContainers":
				assert.Equal(t, tc.expectedValue, totals.ServerlessContainers)
			case "ServerlessFunctions":
				assert.Equal(t, tc.expectedValue, totals.ServerlessFunctions)
			case "VirtualMachines":
				assert.Equal(t, tc.expectedValue, totals.VirtualMachines)
			case "ContainerRegistryImages":
				assert.Equal(t, tc.expectedValue, totals.ContainerRegistryImages)
			}
		})
	}
}
