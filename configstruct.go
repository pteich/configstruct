// Package configstruct provides a parse function to fill a config struct
// with values from cli flags or environment
package configstruct // import "github.com/pteich/configstruct"

import (
	"encoding/json"
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

	// use reflection to deep dive into our struct
	valueRef := reflect.ValueOf(c)
	confType := valueRef.Elem().Type()

	// check if we have a config path in the struct
	if config.file == "" {
		for i := 0; i < confType.NumField(); i++ {
			field := confType.Field(i)
			if field.Tag.Get("config") == "true" {
				// check env
				env := field.Tag.Get("env")
				if env != "" {
					if val, found := os.LookupEnv(env); found {
						config.file = val
						break
					}
				}
				// check cli args
				cli := field.Tag.Get("cli")
				if cli != "" {
					for j, arg := range cliArgs {
						if arg == "-"+cli || arg == "--"+cli {
							if j+1 < len(cliArgs) {
								config.file = cliArgs[j+1]
								break
							}
						}
						if strings.HasPrefix(arg, "-"+cli+"=") || strings.HasPrefix(arg, "--"+cli+"=") {
							parts := strings.SplitN(arg, "=", 2)
							config.file = parts[1]
							break
						}
					}
				}
			}
		}
	}

	// read config file if set
	if config.file != "" {
		err := readConfigFile(c, config)
		if err != nil {
			return err
		}
	}

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
			structSliceValue := &structSliceFlag{
				target: valueRef.Elem().FieldByName(field.Name),
			}

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
				case reflect.Slice:
					if field.Type.Elem().Kind() == reflect.Struct {
						flagSet.Var(structSliceValue, name, usage)
						return nil
					}
					return fmt.Errorf("config cli type %s not implemented", field.Type.String())
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
			if env == "" {
				continue
			}

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
				case reflect.Slice:
					if field.Type.Elem().Kind() != reflect.Struct {
						return fmt.Errorf("config env type %s not implemented", field.Type.String())
					}

					sliceValue, err := decodeStructSliceJSON(envValue, field.Type)
					if err != nil {
						return fmt.Errorf("could not parse env %s for field %s: %w", env, field.Name, err)
					}
					valueRef.Elem().FieldByName(field.Name).Set(sliceValue)
				default:
					return fmt.Errorf("config env type %s not implemented", field.Type.String())
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

type structSliceFlag struct {
	target reflect.Value
	seen   bool
}

func (f *structSliceFlag) String() string {
	if !f.target.IsValid() {
		return ""
	}

	b, err := json.Marshal(f.target.Interface())
	if err != nil {
		return ""
	}

	return string(b)
}

func (f *structSliceFlag) Set(value string) error {
	if !f.target.IsValid() || !f.target.CanSet() {
		return fmt.Errorf("slice field is not settable")
	}

	sliceValue, err := decodeStructSliceJSON(value, f.target.Type())
	if err != nil {
		return err
	}

	if !f.seen {
		f.target.Set(reflect.MakeSlice(f.target.Type(), 0, 0))
		f.seen = true
	}

	f.target.Set(reflect.AppendSlice(f.target, sliceValue))
	return nil
}

func decodeStructSliceJSON(value string, fieldType reflect.Type) (reflect.Value, error) {
	if fieldType.Kind() != reflect.Slice || fieldType.Elem().Kind() != reflect.Struct {
		return reflect.Value{}, fmt.Errorf("type %s is not a slice of structs", fieldType.String())
	}

	sliceTarget := reflect.New(fieldType)
	if err := json.Unmarshal([]byte(value), sliceTarget.Interface()); err == nil {
		return sliceTarget.Elem(), nil
	}

	elemTarget := reflect.New(fieldType.Elem())
	if err := json.Unmarshal([]byte(value), elemTarget.Interface()); err == nil {
		sliceValue := reflect.MakeSlice(fieldType, 1, 1)
		sliceValue.Index(0).Set(elemTarget.Elem())
		return sliceValue, nil
	}

	return reflect.Value{}, fmt.Errorf("value is neither a JSON array nor a JSON object for %s", fieldType.String())
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
	defer f.Close()

	err = yaml.NewDecoder(f).Decode(c)
	if err != nil {
		return fmt.Errorf("could not decode yaml config file %s: %w", cfg.file, err)
	}

	return nil
}

// Save writes the given config struct as YAML to a file
func Save(path string, c interface{}) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("could not create config file %s: %w", path, err)
	}
	defer f.Close()

	encoder := yaml.NewEncoder(f)
	defer encoder.Close()

	err = encoder.Encode(c)
	if err != nil {
		return fmt.Errorf("could not encode yaml config %s: %w", path, err)
	}

	return nil
}
