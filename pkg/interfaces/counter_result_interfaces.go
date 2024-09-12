package interfaces

type CounterResult struct {
	Count                int
	CounterClass         string
	Error                error
	PermissionSuggestion string
}

func (c *CounterResult) Success() bool {
	return c.Error == nil
}
