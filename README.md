# configstruct
Simple Go module to parse a configuration from environment values and CLI flags or arguments using struct tags.
Starting with v1.3.0 there is also support for CLI commands and subcommands

Since v1.5.0 it is possible to define arguments that are also parsed into struct values.

Starting with v1.6.0 parsing a config file is supported (YAML only for now). Use `WithYamlConfig(path)` option to pass
a path to a YAML file that should be parsed. Values passed by flag or env will override values from the config file.

## Usage without commands
```Go
// define a struct with tags for env name, cli flag and usage
type Config struct {
	Filename string `arg:"1" name:"filename" required:"true"`
	Hostname string `env:"CONFIGSTRUCT_HOSTNAME" cli:"hostname" usage:"hostname value"`
	Port     int    `env:"CONFIGSTRUCT_PORT" cli:"port" usage:"listen port"`
	Debug    bool   `env:"CONFIGSTRUCT_DEBUG" cli:"debug" usage:"debug mode"`
}

// create a variable of the struct type and define defaults if needed
conf := Config{
    Hostname: "localhost",
    Port:     8000,
    Debug:    true,
}

// imagine the programm is called like this:
// ./myprogram -hostname=myhost -port=9000 testfile
// the flag values (hostname, port) and argument (filename) are parsed into the struct
// all pre-set defaults are overwritten if a value is provided otherwise it is left as is
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
filename := conf.Filename
if conf.Debug {...}

// cli arguments are also possible

```

## Usage with commands
You can also define "commands" that can be used to execute callback functions. 
The program with global flags and a command `count` should be called like this:
````bash
mycmd -hostname=localhost count -number=2

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

## Share dependencies across commands
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
