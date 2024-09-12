package utils

import (
	"aws-resource-discovery/pkg/interfaces"
	"bytes"
	"testing"

	"github.com/rodaine/table"
	"github.com/stretchr/testify/assert"
)

func TestPrintTotals(t *testing.T) {
	// Redirect stdout to capture the output
	buf := bytes.NewBufferString("")
	table.DefaultWriter = buf
	totals := interfaces.ResourceTotals{
		Buckets:                 10,
		ContainerHosts:          5,
		Databases:               15,
		NonOsDisks:              20,
		ServerlessContainers:    25,
		ServerlessFunctions:     30,
		VirtualMachines:         35,
		ContainerRegistryImages: 40,
	}
	PrintTotals(totals)
	output := buf.String()

	// Check if all resource types are present in the output
	assert.Contains(t, output, "Storage Buckets")
	assert.Contains(t, output, "Container Hosts")
	assert.Contains(t, output, "Databases")
	assert.Contains(t, output, "Non-OS Disks")
	assert.Contains(t, output, "Serverless Containers")
	assert.Contains(t, output, "Serverless Functions")
	assert.Contains(t, output, "Virtual Machines")
	assert.Contains(t, output, "Container Registry Images")

	// Check if all counts are present in the output
	assert.Contains(t, output, "10")
	assert.Contains(t, output, "5")
	assert.Contains(t, output, "15")
	assert.Contains(t, output, "20")
	assert.Contains(t, output, "25")
	assert.Contains(t, output, "30")
	assert.Contains(t, output, "35")
	assert.Contains(t, output, "40")
}
