# configstruct
Simple Go module to parse a configuration from environment values and CLI flags using struct tags.
Starting with v1.3.0 there there is also support for CLI commands and subcommands

## Usage without commands
```Go
// define a struct with tags for env name, cli flag and usage
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

// now parse values from first env and then cli into this var
err := configstruct.Parse(&conf)
if err != nil {...}

// if you prefer env with precedence over cli than use option
err := configstruct.Parse(&conf, configstruct.WithPrecedenceEnv())
if err != nil {...}

// you can also use a special option for the default first cli, then env
err := configstruct.Parse(&conf, configstruct.WithPrecedenceCli())
if err != nil {...}

// after parsing you can pass through you config struct and access values
port := conf.Port
host := conf.Hostname
if conf.Debug {...}
```

## Usage with commands
The program with global flags and a command `count` should be called like this:
````bash
mycmd -hostname localhost count -number 2

```` 

This is the code to model this behaviour:

```Go
// define a struct with tags for env name, cli flag and usage
type RootConfig struct {
	Hostname string `env:"CONFIGSTRUCT_HOSTNAME" cli:"hostname" usage:"hostname value"`
	Debug    bool   `env:"CONFIGSTRUCT_DEBUG" cli:"debug" usage:"debug mode"`
}

type CountConfig struct {
    Number int `cli:"number" usage:"number to count"`
} 

// create a variable of the struct type and define defaults if needed
rootCfg := RootConfig{
    Hostname: "localhost",
    Debug:    true,
}

countCfg := CountConfig {
    Number: 1
}

countCmd := NewCommand("count", &subConfig, func(c *configstruct.Command, cfg interface{}) error {
    cfgValues := cfg.(*CountConfig)
    ...
    return nil
})

cmd := NewCommand("", &rootCfg, func(c *configstruct.Command, cfg interface{}) error {
    cfgValues := cfg.(*RootConfig)
    ...
    return nil
}, subCmd)

err := cmd.ParseAndRun(os.Args)
```

## Share Dependencies accross Commands
It is possible to share dependencies with the command functions `c.SetDependency(name, dep)` and `dep, err := c.GetDependency(name)`.
If you for instance initialize a logger in the root command and register it as dependency every sub-command has
access to it. Keep in mind that dependencies are saved as `interface{}` so you have to take care of asserting the right type.

Taking the example from above further it could look like this:
```Go
type Logger struct {}

countCmd := NewCommand("count", &subConfig, func(c *configstruct.Command, cfg interface{}) error {
    cfgValues := cfg.(*CountConfig)
	loggerDep, err := c.GetDependency("logger")
	if err != nil {
	    return err	
	}
	
	logger := loggerDep.(*Logger)
	
    ...
    return nil
})

cmd := NewCommand("", &rootCfg, func(c *configstruct.Command, cfg interface{}) error {
    cfgValues := cfg.(*RootConfig)
    logger := &Logger{}
    c.SetDependency("logger", logger)
    ...
    return nil
}, subCmd)


```