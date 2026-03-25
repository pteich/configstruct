package configstruct

import (
	"flag"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testConfig struct {
	Hostname   string  `env:"CONFIGSTRUCT_HOSTNAME" cli:"hostname" usage:"hostname value"`
	Port       int     `env:"CONFIGSTRUCT_PORT" cli:"port" cliAlt:"p" usage:"listen port"`
	Debug      bool    `env:"CONFIGSTRUCT_DEBUG" cli:"debug" usage:"debug mode"`
	FloatValue float64 `env:"CONFIGSTRUCT_FLOAT" cli:"floatValue" usage:"float value"`
}

type endpoint struct {
	User string `json:"user" yaml:"user"`
	Pass string `json:"pass" yaml:"pass"`
	URL  string `json:"url" yaml:"url"`
}

type endpointConfig struct {
	Endpoints []endpoint `env:"CONFIGSTRUCT_ENDPOINTS" cli:"endpoints" yaml:"endpoints"`
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

	t.Run("alternative cli field", func(t *testing.T) {
		cliArgs := []string{"command", "-hostname=localhost", "-p=8080", "-debug=true", "-floatValue=100.5"}
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

	t.Run("one required argument", func(t *testing.T) {
		cliArgs := []string{"command", "-hostname=localhost", "-port=8080", "start"}
		flagSet := flag.NewFlagSet(cliArgs[0], flag.ExitOnError)

		conf := struct {
			Hostname string `env:"CONFIGSTRUCT_HOSTNAME" cli:"hostname" usage:"hostname value"`
			Port     int    `env:"CONFIGSTRUCT_PORT" cli:"port" usage:"listen port"`
			Command  string `arg:"1" name:"command" required:"true"`
		}{}

		err := ParseWithFlagSet(flagSet, cliArgs, &conf)
		assert.NoError(t, err)
		assert.Equal(t, "start", conf.Command)
	})

	t.Run("two arguments", func(t *testing.T) {
		cliArgs := []string{"command", "-hostname=localhost", "-port=8080", "start", "myfile"}
		flagSet := flag.NewFlagSet(cliArgs[0], flag.ExitOnError)

		conf := struct {
			Hostname string `env:"CONFIGSTRUCT_HOSTNAME" cli:"hostname" usage:"hostname value"`
			Port     int    `env:"CONFIGSTRUCT_PORT" cli:"port" usage:"listen port"`
			Command  string `arg:"1" name:"command" required:"true"`
			Filename string `arg:"2" name:"filename" required:"true"`
		}{}

		err := ParseWithFlagSet(flagSet, cliArgs, &conf)
		assert.NoError(t, err)
		assert.Equal(t, "start", conf.Command)
		assert.Equal(t, "myfile", conf.Filename)
	})

	t.Run("arguments with defaults", func(t *testing.T) {
		cliArgs := []string{"command", "-hostname=localhost", "-port=8080", "start"}
		flagSet := flag.NewFlagSet(cliArgs[0], flag.ExitOnError)

		conf := struct {
			Hostname string `env:"CONFIGSTRUCT_HOSTNAME" cli:"hostname" usage:"hostname value"`
			Port     int    `env:"CONFIGSTRUCT_PORT" cli:"port" usage:"listen port"`
			Command  string `arg:"1" name:"command" required:"true"`
			Filename string `arg:"2" name:"filename"`
		}{
			Filename: "myfile",
		}

		err := ParseWithFlagSet(flagSet, cliArgs, &conf)
		assert.NoError(t, err)
		assert.Equal(t, "start", conf.Command)
		assert.Equal(t, "myfile", conf.Filename)
	})

	t.Run("required argument missing", func(t *testing.T) {
		cliArgs := []string{"command", "-hostname=localhost", "-port=8080"}
		flagSet := flag.NewFlagSet(cliArgs[0], flag.ExitOnError)

		conf := struct {
			Hostname string `env:"CONFIGSTRUCT_HOSTNAME" cli:"hostname" usage:"hostname value"`
			Port     int    `env:"CONFIGSTRUCT_PORT" cli:"port" usage:"listen port"`
			Command  string `arg:"1" name:"command" required:"true"`
		}{}

		err := ParseWithFlagSet(flagSet, cliArgs, &conf)
		assert.Error(t, err)
	})

	t.Run("save and load yaml", func(t *testing.T) {
		type Config struct {
			Hostname string `yaml:"hostname" cli:"hostname"`
			Port     int    `yaml:"port" cli:"port"`
		}

		conf := Config{Hostname: "test", Port: 1234}
		tmpFile := "test_save.yaml"
		defer os.Remove(tmpFile)

		err := Save(tmpFile, &conf)
		assert.NoError(t, err)

		conf2 := Config{}
		err = ParseWithFlagSet(flag.NewFlagSet("test", flag.ContinueOnError), []string{"test"}, &conf2, WithYamlConfig(tmpFile))
		assert.NoError(t, err)

		assert.Equal(t, conf.Hostname, conf2.Hostname)
		assert.Equal(t, conf.Port, conf2.Port)
	})

	t.Run("dynamic config path via cli", func(t *testing.T) {
		type Config struct {
			ConfigPath string `cli:"config" config:"true"`
			Hostname   string `yaml:"hostname" cli:"hostname"`
		}

		tmpFile := "test_dynamic.yaml"
		defer os.Remove(tmpFile)

		Save(tmpFile, &Config{Hostname: "dynamic-host"})

		conf := Config{}
		cliArgs := []string{"test", "-config", tmpFile}
		err := ParseWithFlagSet(flag.NewFlagSet("test", flag.ContinueOnError), cliArgs, &conf)
		assert.NoError(t, err)
		assert.Equal(t, "dynamic-host", conf.Hostname)
		assert.Equal(t, tmpFile, conf.ConfigPath)
	})

	t.Run("dynamic config path via env", func(t *testing.T) {
		type Config struct {
			ConfigPath string `env:"MY_CONFIG" config:"true"`
			Hostname   string `yaml:"hostname"`
		}

		tmpFile := "test_dynamic_env.yaml"
		defer os.Remove(tmpFile)

		Save(tmpFile, &Config{Hostname: "dynamic-host-env"})

		os.Setenv("MY_CONFIG", tmpFile)
		defer os.Unsetenv("MY_CONFIG")

		conf := Config{}
		err := ParseWithFlagSet(flag.NewFlagSet("test", flag.ContinueOnError), []string{"test"}, &conf)
		assert.NoError(t, err)
		assert.Equal(t, "dynamic-host-env", conf.Hostname)
		assert.Equal(t, tmpFile, conf.ConfigPath)
	})

	t.Run("env json array for struct slice", func(t *testing.T) {
		os.Clearenv()
		os.Setenv("CONFIGSTRUCT_ENDPOINTS", `[{"user":"u1","pass":"p1","url":"https://a"},{"user":"u2","pass":"p2","url":"https://b"}]`)

		cliArgs := []string{"command"}
		flagSet := flag.NewFlagSet(cliArgs[0], flag.ExitOnError)
		conf := endpointConfig{}

		err := ParseWithFlagSet(flagSet, cliArgs, &conf)
		assert.NoError(t, err)
		assert.Len(t, conf.Endpoints, 2)
		assert.Equal(t, "u1", conf.Endpoints[0].User)
		assert.Equal(t, "https://b", conf.Endpoints[1].URL)
	})

	t.Run("env json object for struct slice", func(t *testing.T) {
		os.Clearenv()
		os.Setenv("CONFIGSTRUCT_ENDPOINTS", `{"user":"u1","pass":"p1","url":"https://a"}`)

		cliArgs := []string{"command"}
		flagSet := flag.NewFlagSet(cliArgs[0], flag.ExitOnError)
		conf := endpointConfig{}

		err := ParseWithFlagSet(flagSet, cliArgs, &conf)
		assert.NoError(t, err)
		assert.Len(t, conf.Endpoints, 1)
		assert.Equal(t, "u1", conf.Endpoints[0].User)
		assert.Equal(t, "https://a", conf.Endpoints[0].URL)
	})

	t.Run("cli repeated json objects for struct slice", func(t *testing.T) {
		os.Clearenv()
		cliArgs := []string{
			"command",
			`-endpoints={"user":"u1","pass":"p1","url":"https://a"}`,
			`-endpoints={"user":"u2","pass":"p2","url":"https://b"}`,
		}
		flagSet := flag.NewFlagSet(cliArgs[0], flag.ExitOnError)
		conf := endpointConfig{
			Endpoints: []endpoint{{User: "default", Pass: "default", URL: "https://default"}},
		}

		err := ParseWithFlagSet(flagSet, cliArgs, &conf)
		assert.NoError(t, err)
		assert.Len(t, conf.Endpoints, 2)
		assert.Equal(t, "u1", conf.Endpoints[0].User)
		assert.Equal(t, "u2", conf.Endpoints[1].User)
	})

	t.Run("cli single json array for struct slice", func(t *testing.T) {
		os.Clearenv()
		cliArgs := []string{
			"command",
			`-endpoints=[{"user":"u1","pass":"p1","url":"https://a"},{"user":"u2","pass":"p2","url":"https://b"}]`,
		}
		flagSet := flag.NewFlagSet(cliArgs[0], flag.ExitOnError)
		conf := endpointConfig{}

		err := ParseWithFlagSet(flagSet, cliArgs, &conf)
		assert.NoError(t, err)
		assert.Len(t, conf.Endpoints, 2)
		assert.Equal(t, "u1", conf.Endpoints[0].User)
		assert.Equal(t, "u2", conf.Endpoints[1].User)
	})

	t.Run("default precedence cli over env for struct slice", func(t *testing.T) {
		os.Clearenv()
		os.Setenv("CONFIGSTRUCT_ENDPOINTS", `[{"user":"env","pass":"env","url":"https://env"}]`)

		cliArgs := []string{
			"command",
			`-endpoints={"user":"cli","pass":"cli","url":"https://cli"}`,
		}
		flagSet := flag.NewFlagSet(cliArgs[0], flag.ExitOnError)
		conf := endpointConfig{}

		err := ParseWithFlagSet(flagSet, cliArgs, &conf)
		assert.NoError(t, err)
		assert.Len(t, conf.Endpoints, 1)
		assert.Equal(t, "cli", conf.Endpoints[0].User)
	})

	t.Run("env precedence over cli for struct slice", func(t *testing.T) {
		os.Clearenv()
		os.Setenv("CONFIGSTRUCT_ENDPOINTS", `[{"user":"env","pass":"env","url":"https://env"}]`)

		cliArgs := []string{
			"command",
			`-endpoints={"user":"cli","pass":"cli","url":"https://cli"}`,
		}
		flagSet := flag.NewFlagSet(cliArgs[0], flag.ExitOnError)
		conf := endpointConfig{}

		err := ParseWithFlagSet(flagSet, cliArgs, &conf, WithPrecedenceEnv())
		assert.NoError(t, err)
		assert.Len(t, conf.Endpoints, 1)
		assert.Equal(t, "env", conf.Endpoints[0].User)
	})

	t.Run("yaml struct slice overridden by cli", func(t *testing.T) {
		type Config struct {
			Endpoints []endpoint `yaml:"endpoints" cli:"endpoints"`
		}

		tmpFile := "test_slice_yaml_override.yaml"
		defer os.Remove(tmpFile)

		err := Save(tmpFile, &Config{
			Endpoints: []endpoint{{User: "yaml", Pass: "yaml", URL: "https://yaml"}},
		})
		assert.NoError(t, err)

		conf := Config{}
		cliArgs := []string{
			"command",
			`-endpoints={"user":"cli","pass":"cli","url":"https://cli"}`,
		}
		flagSet := flag.NewFlagSet(cliArgs[0], flag.ExitOnError)
		err = ParseWithFlagSet(flagSet, cliArgs, &conf, WithYamlConfig(tmpFile))
		assert.NoError(t, err)
		assert.Len(t, conf.Endpoints, 1)
		assert.Equal(t, "cli", conf.Endpoints[0].User)
	})

	t.Run("invalid env json returns error for struct slice", func(t *testing.T) {
		os.Clearenv()
		os.Setenv("CONFIGSTRUCT_ENDPOINTS", `{"user":`)

		cliArgs := []string{"command"}
		flagSet := flag.NewFlagSet(cliArgs[0], flag.ExitOnError)
		conf := endpointConfig{}

		err := ParseWithFlagSet(flagSet, cliArgs, &conf)
		assert.Error(t, err)
	})

	t.Run("invalid cli json returns error for struct slice", func(t *testing.T) {
		os.Clearenv()
		cliArgs := []string{"command", `-endpoints={"user":}`}
		flagSet := flag.NewFlagSet(cliArgs[0], flag.ContinueOnError)
		conf := endpointConfig{}

		err := ParseWithFlagSet(flagSet, cliArgs, &conf)
		assert.Error(t, err)
	})

	t.Run("struct slice keeps defaults when no input is set", func(t *testing.T) {
		os.Clearenv()
		cliArgs := []string{"command"}
		flagSet := flag.NewFlagSet(cliArgs[0], flag.ExitOnError)
		conf := endpointConfig{
			Endpoints: []endpoint{{User: "default", Pass: "default", URL: "https://default"}},
		}

		err := ParseWithFlagSet(flagSet, cliArgs, &conf)
		assert.NoError(t, err)
		assert.Len(t, conf.Endpoints, 1)
		assert.Equal(t, "default", conf.Endpoints[0].User)
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
