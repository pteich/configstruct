package configstruct

import (
	"flag"
	"strings"
)

// Command defines a command that consists of a name (empty for root command), a struct that models all
// flags, a function that is executed if the command matches and that gets the config struct as argument
// several sub-commands can be added
type Command struct {
	fs          *flag.FlagSet
	config      interface{}
	f           func(cfg interface{}) error
	subCommands []*Command
}

// NewCommand creates a command that is triggered by the given name in the command line
// all flags are defined by a struct that is parsed and filled with real values
// this struct is then set as argument for the function that is executed if the name matches
func NewCommand(name string, config interface{}, f func(cfg interface{}) error, subCommands ...*Command) *Command {
	fs := flag.NewFlagSet(name, flag.ExitOnError)

	return &Command{
		fs:          fs,
		config:      config,
		f:           f,
		subCommands: subCommands,
	}
}

// ParseAndRun parses the given arguments and executes command functions
func (c *Command) ParseAndRun(args []string, opts ...Option) error {
	err := ParseWithFlagSet(c.fs, args, c.config, opts...)
	if err != nil {
		return err
	}

	if c.fs.Name() == "" || strings.EqualFold(c.fs.Name(), args[0]) {
		err := c.f(c.config)
		if err != nil {
			return err
		}
	}

	args = c.fs.Args()
	if len(args) > 0 {
		for i := range c.subCommands {
			if strings.EqualFold(c.subCommands[i].fs.Name(), args[0]) {
				c.subCommands[i].ParseAndRun(args)
			}
		}
	}

	return nil
}
