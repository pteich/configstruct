package configstruct

import (
	"flag"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
)

// Parse uses a given struct c with tags and parses values from env or cli flags
func Parse(c interface{}) error {

	// use reflection to deep dive into our struct
	valueRef := reflect.ValueOf(c)
	confType := valueRef.Elem().Type()

	// iterate over struct fields for cli flags
	for i := 0; i < confType.NumField(); i++ {
		field := confType.Field(i)
		value := valueRef.Elem().Field(i)
		cli := field.Tag.Get("cli")
		usage := field.Tag.Get("usage")

		if cli != "" {
			switch field.Type.Kind() {
			case reflect.String:
				flag.StringVar(valueRef.Elem().FieldByName(field.Name).Addr().Interface().(*string), cli, value.String(), usage)
			case reflect.Bool:
				flag.BoolVar(valueRef.Elem().FieldByName(field.Name).Addr().Interface().(*bool), cli, value.Bool(), usage)
			case reflect.Int:
				flag.IntVar(valueRef.Elem().FieldByName(field.Name).Addr().Interface().(*int), cli, int(value.Int()), usage)
			default:
				return fmt.Errorf("config cli type %s not implemented", field.Type.Name())
			}
		}
	}

	// parse cli flags
	flag.Parse()

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
			default:
				return fmt.Errorf("config env type %s not implemented", field.Type.Name())
			}
		}
	}

	return nil
}
