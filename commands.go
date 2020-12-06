package configstruct

import (
	"flag"
	"fmt"
	"strings"
)

// CommandFunc is a function that is executed when a command is referenced in a CLI call
type CommandFunc func(cfg interface{}) error

// Command defines a command that consists of a name (empty for root command), a struct that models all
// flags, a function that is executed if the command matches and that gets the config struct as argument
// several sub-commands can be added
type Command struct {
	fs           *flag.FlagSet
	config       interface{}
	f            CommandFunc
	subCommands  []*Command
	rootCommand  *Command
	dependencies map[string]interface{}
}

// NewCommand creates a command that is triggered by the given name in the command line
// all flags are defined by a struct that is parsed and filled with real values
// this struct is then set as argument for the function that is executed if the name matches
func NewCommand(name string, config interface{}, f CommandFunc, subCommands ...*Command) *Command {
	fs := flag.NewFlagSet(name, flag.ExitOnError)

	return &Command{
		fs:           fs,
		config:       config,
		f:            f,
		subCommands:  subCommands,
		dependencies: make(map[string]interface{}),
	}
}

// ParseAndRun parses the given arguments and executes command functions
func (c *Command) ParseAndRun(args []string, opts ...Option) error {
	err := ParseWithFlagSet(c.fs, args, c.config, opts...)
	if err != nil {
		return err
	}

	if c.f != nil && (c.fs.Name() == "" || strings.EqualFold(c.fs.Name(), args[0])) {
		err := c.f(c.config)
		if err != nil {
			return err
		}
	}

	args = c.fs.Args()
	if len(args) > 0 {
		for i := range c.subCommands {
			c.subCommands[i].rootCommand = c
			if strings.EqualFold(c.subCommands[i].fs.Name(), args[0]) {
				c.subCommands[i].ParseAndRun(args)
			}
		}
	}

	return nil
}

// SetDependency saves a dependency referenced by a name for subcommands
func (c *Command) SetDependency(name string, dep interface{}) {
	c.dependencies[name] = dep
}

// GetDependency gets a previous saved dependency in this command or any of the parent commands in the chain
func (c *Command) GetDependency(name string) (interface{}, error) {
	depValue, found := c.dependencies[name]
	if found {
		return depValue, nil
	}

	if c.rootCommand == nil {
		return nil, fmt.Errorf("dependency %s not found", name)
	}

	return c.rootCommand.GetDependency(name)
}
