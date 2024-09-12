package interfaces

import (
	"context"
	"aws-resource-discovery/pkg/config"

	"github.com/aws/aws-sdk-go-v2/aws"
)

type ResourceScanner interface {
	Call()
	ScanResources()
	CreateSession(ctx context.Context) aws.Config
}

type ResourceTotals struct {
	Buckets                 int
	ContainerHosts          int
	ContainerRegistryImages int
	Databases               int
	NonOsDisks              int
	ServerlessContainers    int
	ServerlessFunctions     int
	VirtualMachines         int
}

type Scanner interface {
	ScanSingleAccount(ctx context.Context, config config.Config, logger Logger) ScanResult
	ScanOrganization(ctx context.Context, config config.Config, logger Logger) ScanResult
}
