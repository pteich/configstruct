package configstruct

type config struct {
	precedenceEnv bool
	file          string
}

// Option is a config setting function
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

// WithYamlConfig sets the path to a yaml config file
func WithYamlConfig(path string) Option {
	return func(c *config) {
		c.file = path
	}
}
