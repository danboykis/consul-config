package consulconfig

import (
	"encoding/json"
	"fmt"
	"github.com/hashicorp/consul/api"
	"reflect"
	"strconv"
	"strings"
)

func ReadConsulConfig[C any](conf *C, prefix, token, address string) error {

	cconfig := &api.Config{Token: token, Address: address}
	c, err := api.NewClient(cconfig)
	if err != nil {
		return err
	}
	kv := c.KV()
	kvpairs, _, err := kv.List(prefix, &api.QueryOptions{})
	if err != nil {
		return err
	}

	cTree := newConfigTree()
	for _, p := range kvpairs {
		cTree.addNode(strings.TrimPrefix(p.Key, prefix), string(p.Value))
	}

	populateConfig(cTree, reflect.ValueOf(conf))
	return nil
}

func unwrap(v interface{}) reflect.Value {
	switch v.(type) {
	case reflect.Value:
		return v.(reflect.Value)
	default:
		return reflect.ValueOf(v)
	}
}

func derefValue(v interface{}) (reflect.Value, error) {

	val := unwrap(v)

	if val.Kind() != reflect.Ptr || val.IsNil() {
		return reflect.Value{}, fmt.Errorf("invalid %s val %+v", val.Kind(), v)
	}

	for val.Type().Kind() == reflect.Ptr {
		val = val.Elem()
	}
	return val, nil
}

func PopulateConfig(configMap map[string]string, ptr interface{}) error {
	if _, err := populateConfig(newConfigTreeFromMap(configMap), ptr); err != nil {
		return err
	}
	return nil
}

func populateConfig(confTree *configTree, ptr interface{}) (reflect.Value, error) {
	v, derr := derefValue(ptr)
	if derr != nil {
		return reflect.Value{}, derr
	}
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		tag := v.Type().Field(i).Tag.Get("config")
		defaultVal := ""
		if tagAndDefault := strings.Split(tag, ","); len(tagAndDefault) == 2 {
			tag = tagAndDefault[0]
			defaultVal = tagAndDefault[1]
		}
		switch field.Kind() {
		case reflect.Map:
			if jsonValue, ok := confTree.getString(tag); ok {
				stringMap := make(map[string]string)
				var m map[string]interface{}
				if err := json.Unmarshal([]byte(jsonValue), &m); err == nil {
					for k, v := range m {
						switch val := v.(type) {
						case string:
							stringMap[k] = val
						default:
							stringMap[k] = fmt.Sprintf("%s", v)
						}
					}
					field.Set(reflect.ValueOf(stringMap))
				}
			}
		case reflect.Struct:
			if st, exists := confTree.getSubTree(tag); exists {
				if v, err := populateConfig(st, field.Addr()); err != nil {
					return reflect.Value{}, err
				} else {
					field.Set(v)
				}
			}
		case reflect.String:
			if v, ok := confTree.getString(tag); ok {
				field.Set(reflect.ValueOf(v))
			} else if defaultVal != "" {
				field.Set(reflect.ValueOf(defaultVal))
			}
		case reflect.Int:
			if v, ok := confTree.getString(tag); ok {
				if intVal, perr := strconv.Atoi(v); perr == nil {
					field.Set(reflect.ValueOf(intVal))
				}
			}
		case reflect.Int64:
			if v, ok := confTree.getString(tag); ok {
				if int64Val, perr := strconv.ParseInt(v, 10, 64); perr == nil {
					field.Set(reflect.ValueOf(int64Val))
				}
			}
		case reflect.Float64:
			if v, ok := confTree.getString(tag); ok {
				if f64Val, perr := strconv.ParseFloat(v, 64); perr == nil {
					field.Set(reflect.ValueOf(f64Val))
				}
			}
		case reflect.Bool:
			if v, ok := confTree.getString(tag); ok {
				if boolVal, _ := strconv.ParseBool(v); boolVal {
					field.Set(reflect.ValueOf(boolVal))
				} else {
					field.Set(reflect.ValueOf(false))
				}
			}
		case reflect.Slice:
			switch field.Type().Elem().Kind() {
			case reflect.String:
				if v, ok := confTree.getString(tag); ok {
					var vHolder []string
					if err := json.Unmarshal([]byte(v), &vHolder); err == nil {
						field.Set(reflect.ValueOf(vHolder))
					}
				}
			}
		}
	}
	return v, nil
}
