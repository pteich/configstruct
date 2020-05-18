package configstruct

type config struct {
	precedenceEnv bool
}

// Options is an config setting function
type Option func(c *config)

// WithPrecedenceEnv enabled precedence of ENV values over cli
func WithPrecedenceEnv() Option {
	return func(c *config) {
		c.precedenceEnv = true
	}
}

// WithPrecedenceCli enabled precedence of cli over ENV values (default)
func WithPrecedenceCli() Option {
	return func(c *config) {
		c.precedenceEnv = false
	}
}
