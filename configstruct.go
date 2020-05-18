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
)

// Parse uses a given struct c with tags and parses values from env or cli flags, it uses the default FlagSet and os.Args
func Parse(c interface{}, opts ...Option) error {
	return ParseWithFlagSet(flag.CommandLine, os.Args, c, opts...)
}

// ParseWithFlagSet can use a specific FlagSet and args slice to parse data from
func ParseWithFlagSet(flagSet *flag.FlagSet, cliArgs []string, c interface{}, opts ...Option) error {
	config := config{
		precedenceEnv: false,
	}
	for _, opt := range opts {
		opt(&config)
	}

	// use reflection to deep dive into our struct
	valueRef := reflect.ValueOf(c)
	confType := valueRef.Elem().Type()

	// parse cli flags
	parseCli := func() error {
		// iterate over struct fields for cli flags
		for i := 0; i < confType.NumField(); i++ {
			field := confType.Field(i)
			value := valueRef.Elem().Field(i)
			cli := field.Tag.Get("cli")
			usage := field.Tag.Get("usage")

			if cli != "" {
				switch field.Type.Kind() {
				case reflect.String:
					flagSet.StringVar(valueRef.Elem().FieldByName(field.Name).Addr().Interface().(*string), cli, value.String(), usage)
				case reflect.Bool:
					flagSet.BoolVar(valueRef.Elem().FieldByName(field.Name).Addr().Interface().(*bool), cli, value.Bool(), usage)
				case reflect.Int:
					flagSet.IntVar(valueRef.Elem().FieldByName(field.Name).Addr().Interface().(*int), cli, int(value.Int()), usage)
				case reflect.Float64:
					flagSet.Float64Var(valueRef.Elem().FieldByName(field.Name).Addr().Interface().(*float64), cli, value.Float(), usage)
				default:
					return fmt.Errorf("config cli type %s not implemented", field.Type.Name())
				}
			}
		}

		flagSet.Parse(cliArgs[1:])
		return nil
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

	return nil
}
