package interfaces

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/organizations"
	"github.com/aws/aws-sdk-go-v2/service/organizations/types"
)

type OrganizationsClient interface {
	DescribeAccount(ctx context.Context, input *organizations.DescribeAccountInput, opts ...func(*organizations.Options)) (*organizations.DescribeAccountOutput, error)
	ListAccounts(ctx context.Context, input *organizations.ListAccountsInput, opts ...func(*organizations.Options)) (*organizations.ListAccountsOutput, error)
}

type OrgDetector interface {
	ListAccounts() []types.Account
	SuggestPermissions(msg string)
}

type OrgAccountsManager interface {
	ListAccounts() ([]types.Account, error)
}

type OrgScanner interface {
	Call()
}

type AccountFilter interface {
	FilterActiveAccounts() []types.Account
}
