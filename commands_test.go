package configstruct

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

type rootCmdConfig struct {
	Hostname   string  `env:"CONFIGSTRUCT_HOSTNAME" cli:"hostname" usage:"hostname value"`
	Port       int     `env:"CONFIGSTRUCT_PORT" cli:"port" usage:"listen port"`
	Debug      bool    `env:"CONFIGSTRUCT_DEBUG" cli:"debug" usage:"debug mode"`
	FloatValue float64 `env:"CONFIGSTRUCT_FLOAT" cli:"floatValue" usage:"float value"`
}

type testStruct struct {
	testValue string
}

type subCmdConfig struct {
	Number int `cli:"number" usage:"number to count"`
}

func TestCommand_ParseAndRun(t *testing.T) {
	args := []string{"cliName", "-hostname", "localhost", "math", "count", "-number", "2"}

	var rootConfig rootCmdConfig
	var countConfig subCmdConfig

	countCmd := NewCommand("count", &countConfig, func(cmd *Command, cfg interface{}) error {
		cfgValues := cfg.(*subCmdConfig)
		t.Log("count command", cfgValues.Number)
		return nil
	})

	mathCmd := NewCommand("math", nil, nil, countCmd)

	cmd := NewCommand("", &rootConfig, func(cmd *Command, cfg interface{}) error {
		cfgValues := cfg.(*rootCmdConfig)
		t.Log("root command", cfgValues.Hostname)
		return nil
	}, mathCmd)

	err := cmd.ParseAndRun(args)
	if err != nil {
		t.Errorf("error should be nil but is %v", err)
	}
}

func TestCommand_Dependencies(t *testing.T) {
	c := NewCommand("testCmd", nil, nil)

	test := &testStruct{testValue: "test"}

	c.SetDependency("test", test)

	testReturn, err := c.GetDependency("test")
	assert.NoError(t, err)

	assert.IsType(t, &testStruct{}, testReturn)
	assert.Equal(t, "test", testReturn.(*testStruct).testValue)
}
