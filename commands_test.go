package configstruct

import (
	"testing"
)

type rootCmdConfig struct {
	Hostname   string  `env:"CONFIGSTRUCT_HOSTNAME" cli:"hostname" usage:"hostname value"`
	Port       int     `env:"CONFIGSTRUCT_PORT" cli:"port" usage:"listen port"`
	Debug      bool    `env:"CONFIGSTRUCT_DEBUG" cli:"debug" usage:"debug mode"`
	FloatValue float64 `env:"CONFIGSTRUCT_FLOAT" cli:"floatValue" usage:"float value"`
}

type subCmdConfig struct {
	Number int `cli:"number" usage:"number to count"`
}

func TestCommand_ParseAndRun(t *testing.T) {
	args := []string{"cliName", "-hostname", "localhost", "count", "-number", "2"}

	var rootConfig rootCmdConfig
	var subConfig subCmdConfig

	subCmd := NewCommand("count", &subConfig, func(cfg interface{}) error {
		cfgValues := cfg.(*subCmdConfig)
		t.Log("sub command", cfgValues.Number)
		return nil
	})

	cmd := NewCommand("", &rootConfig, func(cfg interface{}) error {
		cfgValues := cfg.(*rootCmdConfig)
		t.Log("root command", cfgValues.Hostname)
		return nil
	}, subCmd)

	err := cmd.ParseAndRun(args)
	if err != nil {
		t.Errorf("error should be nil but is %v", err)
	}
}
