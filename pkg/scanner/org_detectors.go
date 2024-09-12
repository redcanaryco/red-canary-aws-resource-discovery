package scanner

import (
	"aws-resource-discovery/pkg/interfaces"
	"context"

	"github.com/aws/aws-sdk-go-v2/service/organizations"
	"github.com/aws/aws-sdk-go-v2/service/organizations/types"
)

type orgDetector struct {
	Client interfaces.OrganizationsClient
	Logger interfaces.Logger
}

func NewOrgDetector(client interfaces.OrganizationsClient, logger interfaces.Logger) interfaces.OrgDetector {
	return &orgDetector{
		Client: client,
		Logger: logger,
	}
}

func (d *orgDetector) ListAccounts() []types.Account {
	accounts := []types.Account{}
	var nextToken *string

	for {
		resp, err := d.Client.ListAccounts(context.TODO(), &organizations.ListAccountsInput{
			NextToken: nextToken,
		})
		if err != nil {
			switch {
			case isAWSOrganizationsNotInUseException(err):
				return nil
			case isAccessDeniedException(err):
				d.SuggestPermissions("Unable to list accounts in the organization.")
				return nil
			default:
				d.SuggestPermissions("Unable to determine whether the given account belongs to an organization.")
				return nil
			}
		}
		accounts = append(accounts, resp.Accounts...)
		if resp.NextToken == nil {
			break
		}
		nextToken = resp.NextToken
	}
	return accounts
}

func (d *orgDetector) SuggestPermissions(msg string) {
	d.Logger.Logf(`%s

To scan an organization, the provided credentials must have the following permissions:
  - organizations:ListAccounts
`, msg)
}

func isAWSOrganizationsNotInUseException(err error) bool {
	_, ok := err.(*types.AWSOrganizationsNotInUseException)
	return ok
}

func isAccessDeniedException(err error) bool {
	_, ok := err.(*types.AccessDeniedException)
	return ok
}
