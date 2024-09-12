package config

// Config defines the configuration for the scanner.
type Config struct {
	RoleArn         string
	AccountId       string
	Region          string
	RoleName        string
	Trail           bool
	ExcludeAccounts []string
}
