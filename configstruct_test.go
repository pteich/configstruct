package configstruct

import (
	"flag"
	"fmt"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

type testConfig struct {
	Hostname   string  `env:"CONFIGSTRUCT_HOSTNAME" cli:"hostname" usage:"hostname value"`
	Port       int     `env:"CONFIGSTRUCT_PORT" cli:"port" usage:"listen port"`
	Debug      bool    `env:"CONFIGSTRUCT_DEBUG" cli:"debug" usage:"debug mode"`
	FloatValue float64 `env:"CONFIGSTRUCT_FLOAT" cli:"floatValue" usage:"float value"`
}

func TestParse(t *testing.T) {
	t.Run("valid cli fields", func(t *testing.T) {
		cliArgs := []string{"command", "-hostname=localhost", "-port=8080", "-debug=true", "-floatValue=100.5"}
		flagSet := flag.NewFlagSet(cliArgs[0], flag.ExitOnError)

		conf := testConfig{}

		err := ParseWithFlagSet(flagSet, cliArgs, &conf)
		assert.NoError(t, err)

		assert.Equal(t, 8080, conf.Port)
		assert.Equal(t, "localhost", conf.Hostname)
		assert.True(t, conf.Debug)
		assert.Equal(t, 100.5, conf.FloatValue)
	})

	t.Run("set using env fields", func(t *testing.T) {
		cliArgs := []string{"command", "-hostname=localhost", "-port=8080", "-debug=true", "-floatValue=100.5"}
		flagSet := flag.NewFlagSet(cliArgs[0], flag.ExitOnError)

		os.Clearenv()
		os.Setenv("CONFIGSTRUCT_HOSTNAME", "myhost")
		os.Setenv("CONFIGSTRUCT_PORT", "9000")
		os.Setenv("CONFIGSTRUCT_DEBUG", "true")
		os.Setenv("CONFIGSTRUCT_FLOAT", "2.5")

		conf := testConfig{}

		err := ParseWithFlagSet(flagSet, cliArgs, &conf, WithPrecedenceEnv())
		assert.NoError(t, err)

		assert.Equal(t, 9000, conf.Port)
		assert.Equal(t, "myhost", conf.Hostname)
		assert.True(t, conf.Debug)
		assert.Equal(t, 2.5, conf.FloatValue)
	})

	t.Run("overwrite cli flags with env fields", func(t *testing.T) {
		cliArgs := []string{"command", "-hostname=server", "-port=8000", "-debug=true", "-floatValue=100.5"}
		flagSet := flag.NewFlagSet(cliArgs[0], flag.ExitOnError)

		os.Clearenv()
		os.Setenv("CONFIGSTRUCT_HOSTNAME", "myhost")
		os.Setenv("CONFIGSTRUCT_PORT", "9000")

		conf := testConfig{}

		err := ParseWithFlagSet(flagSet, cliArgs, &conf, WithPrecedenceEnv())
		assert.NoError(t, err)

		assert.Equal(t, 9000, conf.Port)
		assert.Equal(t, "myhost", conf.Hostname)
		assert.True(t, conf.Debug)
	})

	t.Run("cli with defaults", func(t *testing.T) {
		os.Args = []string{"command", "-hostname", "myhost"}
		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

		os.Clearenv()

		conf := testConfig{
			Hostname:   "localhost",
			Port:       8000,
			Debug:      true,
			FloatValue: 300.1,
		}

		err := Parse(&conf)
		assert.NoError(t, err)

		assert.Equal(t, 8000, conf.Port)
		assert.Equal(t, "myhost", conf.Hostname)
		assert.True(t, conf.Debug)
		assert.Equal(t, 300.1, conf.FloatValue)
	})

	t.Run("not implemented types", func(t *testing.T) {
		cliArgs := []string{"command", "-hostname=localhost", "-port=8080", "-debug=true", "-floatValue=100.5"}
		flagSet := flag.NewFlagSet(cliArgs[0], flag.ExitOnError)

		conf := struct {
			Hostname string `env:"CONFIGSTRUCT_HOSTNAME" cli:"hostname" usage:"hostname value"`
			Port     int64  `env:"CONFIGSTRUCT_PORT" cli:"port" usage:"listen port"`
		}{}

		err := ParseWithFlagSet(flagSet, cliArgs, &conf)
		assert.Error(t, err)
	})

}

// Example for using `configstruct` with default values.
func ExampleParse() {
	// define a struct with tags
	type Config struct {
		Hostname string `env:"CONFIGSTRUCT_HOSTNAME" cli:"hostname" usage:"hostname value"`
		Port     int    `env:"CONFIGSTRUCT_PORT" cli:"port" usage:"listen port"`
		Debug    bool   `env:"CONFIGSTRUCT_DEBUG" cli:"debug" usage:"debug mode"`
	}

	// create a variable of the struct type and define defaults if needed
	conf := testConfig{
		Hostname: "localhost",
		Port:     8000,
		Debug:    true,
	}

	// now parse values from first cli flags and then env into this var
	err := Parse(&conf)
	if err != nil {
		fmt.Printf("can't parse config %s", err)
	}
}
