// Package configstruct provides a parse function to fill a config struct
// with values from cli flags or environment
package configstruct // import "github.com/pteich/configstruct"

import (
	"flag"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// Parse uses a given struct c with tags and parses values from env or cli flags, it uses the default FlagSet and os.Args
func Parse(c interface{}, opts ...Option) error {
	return ParseWithFlagSet(flag.NewFlagSet(os.Args[0], flag.ContinueOnError), os.Args, c, opts...)
}

// ParseWithFlagSet can use a specific FlagSet and args slice to parse data from
func ParseWithFlagSet(flagSet *flag.FlagSet, cliArgs []string, c interface{}, opts ...Option) error {
	config := config{
		precedenceEnv: false,
	}
	for _, opt := range opts {
		opt(&config)
	}

	if c == nil {
		flagSet.Parse(cliArgs[1:])
		return nil
	}

	// read config file if set
	if config.file != "" {
		err := readConfigFile(c, config)
		if err != nil {
			return err
		}
	}

	// use reflection to deep dive into our struct
	valueRef := reflect.ValueOf(c)
	confType := valueRef.Elem().Type()

	// parse arguments
	parseArgs := func() error {
		// iterate over struct fields for arg flags
		for i := 0; i < confType.NumField(); i++ {
			field := confType.Field(i)
			value := valueRef.Elem().Field(i)
			name := field.Tag.Get("name")

			required := field.Tag.Get("required") == "true"
			arg, err := strconv.Atoi(field.Tag.Get("arg"))
			if err != nil {
				arg = -1
			}

			if arg > 0 {
				argVal := flagSet.Arg(arg - 1)
				if required && argVal == "" {
					flagSet.Usage()
					return fmt.Errorf("argument %s is required", name)
				}

				if argVal != "" {
					value.Set(reflect.ValueOf(argVal))
				}
			}
		}

		return nil
	}

	// parse cli flags
	parseCli := func() error {
		// iterate over struct fields for cli flags
		for i := 0; i < confType.NumField(); i++ {
			field := confType.Field(i)
			value := valueRef.Elem().Field(i)
			cli := field.Tag.Get("cli")
			cliAlt := field.Tag.Get("cliAlt")
			usage := field.Tag.Get("usage")

			setFlag := func(name string) error {
				switch field.Type.Kind() {
				case reflect.String:
					flagSet.StringVar(valueRef.Elem().FieldByName(field.Name).Addr().Interface().(*string), name, value.String(), usage)
				case reflect.Bool:
					flagSet.BoolVar(valueRef.Elem().FieldByName(field.Name).Addr().Interface().(*bool), name, value.Bool(), usage)
				case reflect.Int:
					flagSet.IntVar(valueRef.Elem().FieldByName(field.Name).Addr().Interface().(*int), name, int(value.Int()), usage)
				case reflect.Float64:
					flagSet.Float64Var(valueRef.Elem().FieldByName(field.Name).Addr().Interface().(*float64), name, value.Float(), usage)
				default:
					return fmt.Errorf("config cli type %s not implemented", field.Type.Kind())
				}

				return nil
			}

			if cli != "" {
				err := setFlag(cli)
				if err != nil {
					return err
				}
			}
			if cliAlt != "" {
				err := setFlag(cliAlt)
				if err != nil {
					return err
				}
			}
		}

		return flagSet.Parse(cliArgs[1:])
	}

	parseEnv := func() error {
		// iterate over struct fields for env values
		for i := 0; i < confType.NumField(); i++ {
			field := confType.Field(i)
			env := field.Tag.Get("env")

			envValue, found := os.LookupEnv(env)
			if found {
				switch field.Type.Kind() {
				case reflect.String:
					valueRef.Elem().FieldByName(field.Name).SetString(envValue)
				case reflect.Bool:
					valueRef.Elem().FieldByName(field.Name).SetBool(false)
					if strings.EqualFold(envValue, "true") {
						valueRef.Elem().FieldByName(field.Name).SetBool(true)
					}
				case reflect.Int:
					value, err := strconv.ParseInt(envValue, 0, 64)
					if err == nil {
						valueRef.Elem().FieldByName(field.Name).SetInt(value)
					}
				case reflect.Float64:
					value, err := strconv.ParseFloat(envValue, 64)
					if err == nil {
						valueRef.Elem().FieldByName(field.Name).SetFloat(value)
					}
				default:
					return fmt.Errorf("config env type %s not implemented", field.Type.Name())
				}
			}
		}

		return nil
	}

	var err error
	if config.precedenceEnv {
		err = parseCli()
		if err != nil {
			return err
		}
		err = parseEnv()
		if err != nil {
			return err
		}
		err = parseArgs()
		if err != nil {
			return err
		}

		return nil
	}

	err = parseEnv()
	if err != nil {
		return err
	}
	err = parseCli()
	if err != nil {
		return err
	}
	err = parseArgs()
	if err != nil {
		return err
	}

	return nil
}

type structFlag struct {
	name         string
	description  string
	defaultValue interface{}
}

func getStructFlags(c interface{}) []structFlag {
	f := make([]structFlag, 0)

	valueRef := reflect.ValueOf(c)
	confType := valueRef.Elem().Type()

	for i := 0; i < confType.NumField(); i++ {
		field := confType.Field(i)
		value := valueRef.Elem().Field(i)
		cli := field.Tag.Get("cli")
		usage := field.Tag.Get("usage")

		f = append(f, structFlag{
			name:         cli,
			description:  usage,
			defaultValue: value,
		})
	}

	return f
}

func readConfigFile(c interface{}, cfg config) error {
	f, err := os.Open(cfg.file)
	if err != nil {
		return fmt.Errorf("could not open config file %s: %w", cfg.file, err)
	}

	err = yaml.NewDecoder(f).Decode(c)
	if err != nil {
		return fmt.Errorf("could not decode yaml config file %s: %w", cfg.file, err)
	}

	return nil
}
