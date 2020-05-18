# configstruct
Simple Go module to parse a configuration from environment and cli flags using struct tags.

Usage
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
