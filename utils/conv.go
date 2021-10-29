package utils

import (
	"encoding/json"
	"reflect"
	"strconv"
	"strings"
)

func ConvValue(target interface{}, tmp string) error {
	if target == nil {
		return nil
	}

	rv := reflect.ValueOf(target)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}

	switch rv.Kind() {
	case reflect.String:
		rv.SetString(tmp)
	case reflect.Bool:
		v, err := strconv.ParseBool(tmp)
		rv.SetBool(v)
		return err
	case reflect.Float32, reflect.Float64:
		v, err := strconv.ParseFloat(tmp, 64)
		rv.SetFloat(v)
		return err
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		v, err := strconv.ParseInt(tmp, 10, 64)
		rv.SetInt(v)
		return err
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		v, err := strconv.ParseUint(tmp, 10, 64)
		rv.SetUint(v)
		return err
	}
	return nil
}

func GetJsonTagName(jsonTag string) string {
	if jsonTag == "" || jsonTag == "-" {
		return ""
	}
	ss := strings.Split(jsonTag, ",")
	return ss[0]
}

func IsBaseType(typ string) bool {
	switch typ {
	case "bool", "string":
		return true
	case "float32", "float64":
		return true
	case "int", "int8", "int16", "int32", "int64":
		return true
	case "uint", "uint8", "uint16", "uint32", "uint64":
		return true
	}
	return false
}

func JsonPretty(obj interface{}) string {
	ds, _ := json.MarshalIndent(obj, "", "  ")
	return string(ds)
}
