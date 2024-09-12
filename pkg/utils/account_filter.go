package utils

import (
	"aws-resource-discovery/pkg/interfaces"
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/organizations"
	"github.com/aws/aws-sdk-go-v2/service/organizations/types"
)

type accountFilter struct {
	OrgAccounts     []types.Account
	OrgClient       interfaces.OrganizationsClient
	Logger          interfaces.Logger
	ExcludeAccounts []string
}

func NewAccountFilter(orgAccounts []types.Account, orgClient interfaces.OrganizationsClient, logger interfaces.Logger, excludeAccounts []string) interfaces.AccountFilter {
	return &accountFilter{
		OrgAccounts:     orgAccounts,
		OrgClient:       orgClient,
		Logger:          logger,
		ExcludeAccounts: excludeAccounts,
	}
}

func (af *accountFilter) FilterActiveAccounts() []types.Account {
	activeAccounts := []types.Account{}
	for _, account := range af.OrgAccounts {
		if af.isExcludedAccount(*account.Id) {
			af.Logger.Logf("Skipping excluded account: %s", *account.Id)
			continue
		}
		if af.isAccountActive(*account.Id) {
			activeAccounts = append(activeAccounts, account)
		} else {
			af.Logger.Logf("Skipping suspended account: %s", *account.Id)
		}
	}
	return activeAccounts
}

func (af *accountFilter) isAccountActive(accountId string) bool {
    input := &organizations.DescribeAccountInput{
        AccountId: aws.String(accountId),
    }

    result, err := af.OrgClient.DescribeAccount(context.Background(), input)
    if err != nil {
        af.Logger.Logf("Failed to describe account %s: %v", accountId, err)
        return false
    }

    return result.Account.Status == types.AccountStatusActive
}

func (af *accountFilter) isExcludedAccount(accountId string) bool {
	for _, excludedAccount := range af.ExcludeAccounts {
		if accountId == excludedAccount {
			return true
		}
	}
	return false
}
