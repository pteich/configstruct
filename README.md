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

// now parse values from first cli flags and then env into this var
err := configstruct.Parse(&conf)
if err != nil {...}


```
