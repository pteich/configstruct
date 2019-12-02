package configstruct

import (
	"flag"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

type testConfig struct {
	Hostname string `env:"CONFIGSTRUCT_HOSTNAME" cli:"hostname" usage:"hostname value"`
	Port     int    `env:"CONFIGSTRUCT_PORT" cli:"port" usage:"listen port"`
	Debug    bool   `env:"CONFIGSTRUCT_DEBUG" cli:"debug" usage:"debug mode"`
}

func TestParse(t *testing.T) {

	t.Run("valid cli fields", func(t *testing.T) {
		os.Args = []string{"command", "-hostname", "localhost", "-port", "8080", "-debug", "true"}
		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

		conf := testConfig{}

		err := Parse(&conf)
		assert.NoError(t, err)

		assert.Equal(t, 8080, conf.Port)
		assert.Equal(t, "localhost", conf.Hostname)
		assert.True(t, conf.Debug)

	})

	t.Run("set using env fields", func(t *testing.T) {
		os.Clearenv()
		os.Setenv("CONFIGSTRUCT_HOSTNAME", "myhost")
		os.Setenv("CONFIGSTRUCT_PORT", "9000")
		os.Setenv("CONFIGSTRUCT_DEBUG", "true")

		conf := testConfig{}

		err := Parse(&conf)
		assert.NoError(t, err)

		assert.Equal(t, 9000, conf.Port)
		assert.Equal(t, "myhost", conf.Hostname)
		assert.True(t, conf.Debug)

	})

	t.Run("overwrite cli flags with env fields", func(t *testing.T) {
		os.Args = []string{"command", "-hostname", "localhost", "-port", "8080", "-debug", "true"}
		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

		os.Clearenv()
		os.Setenv("CONFIGSTRUCT_HOSTNAME", "myhost")
		os.Setenv("CONFIGSTRUCT_PORT", "9000")

		conf := testConfig{}

		err := Parse(&conf)
		assert.NoError(t, err)

		assert.Equal(t, 9000, conf.Port)
		assert.Equal(t, "myhost", conf.Hostname)
		assert.True(t, conf.Debug)
	})

	t.Run("cli with defaults", func(t *testing.T) {
		os.Args = []string{"command", "-hostname", "myhost"}
		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

		conf := testConfig{
			Hostname: "localhost",
			Port:     8000,
			Debug:    true,
		}

		err := Parse(&conf)
		assert.NoError(t, err)

		assert.Equal(t, 8000, conf.Port)
		assert.Equal(t, "myhost", conf.Hostname)
		assert.True(t, conf.Debug)
	})
}
