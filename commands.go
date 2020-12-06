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
	fs.Usage = func() {
		if name == "" {
			fmt.Fprintf(fs.Output(), "Usage:\n")
		} else {
			fmt.Fprintf(fs.Output(), "Usage of %s:\n", name)
		}
		fs.PrintDefaults()

		if len(subCommands) > 0 {
			fmt.Fprintf(fs.Output(), "Available Subcommands: ")
			for i := range subCommands {
				fmt.Fprintf(fs.Output(), subCommands[i].fs.Name()+" ")
			}
			fmt.Fprintf(fs.Output(), "\n")
		}
	}

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
	if len(c.subCommands) > 0 && len(args) == 0 {
		c.fs.Usage()
		return nil
	}

	if len(args) > 0 {
		cmdFound := false
		for i := range c.subCommands {
			if strings.EqualFold(c.subCommands[i].fs.Name(), args[0]) {
				c.subCommands[i].rootCommand = c
				cmdFound = true
				c.subCommands[i].ParseAndRun(args)
			}
		}
		if !cmdFound {
			fmt.Fprintf(c.fs.Output(), "Command '%s' not defined\n\n", args[0])
			c.fs.Usage()
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
