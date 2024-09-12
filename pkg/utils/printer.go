package utils

import (
    "aws-resource-discovery/pkg/interfaces"
    "github.com/fatih/color"
    "github.com/rodaine/table"
)

func PrintTotals(totals interfaces.ResourceTotals) {
    headerFmt := color.New(color.FgGreen, color.Underline).SprintfFunc()
    columnFmt := color.New(color.FgYellow).SprintfFunc()
    tbl := table.New("ResourceType", "Count").WithPadding(3)
    tbl.WithHeaderFormatter(headerFmt).WithFirstColumnFormatter(columnFmt)
    tbl.AddRow("Storage Buckets", totals.Buckets)
    tbl.AddRow("Container Hosts", totals.ContainerHosts)
    tbl.AddRow("Databases", totals.Databases)
    tbl.AddRow("Non-OS Disks", totals.NonOsDisks)
    tbl.AddRow("Serverless Containers", totals.ServerlessContainers)
    tbl.AddRow("Serverless Functions", totals.ServerlessFunctions)
    tbl.AddRow("Virtual Machines", totals.VirtualMachines)
    tbl.AddRow("Container Registry Images", totals.ContainerRegistryImages)
    tbl.Print()
}