package util

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/globalsign/mgo/bson"
)

func Bson2struct(m bson.M, s interface{}) (err error) {
	rt := reflect.TypeOf(s)
	rv := reflect.ValueOf(s)

	for rt.Kind() == reflect.Ptr {
		rt = rt.Elem()
		rv = rv.Elem()
	}

	if rt.Kind() != reflect.Struct {
		err = fmt.Errorf("not struct %v", rt.Kind())
		return
	}

	for i := 0; i < rt.NumField(); i++ {
		sf := rt.Field(i)
		k := sf.Tag.Get("json")
		k = strings.ToLower(k)
		kk := strings.Split(k, ",")
		if len(kk) < 1 {
			continue
		}
		key := kk[0]
		if key == "-" {
			continue
		}
		value := rv.Field(i)

		switch sf.Type.Kind() {
		case reflect.Bool:
			if v, ok := m[key].(bool); ok {
				value.SetBool(v)
			}
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if v, ok := m[key].(int64); ok {
				value.SetInt(v)
			}
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			if v, ok := m[key].(int64); ok {
				if v < 0 {
					fmt.Println("小于0", v, sf.Name)
					v = 0
				}
				value.SetUint(uint64(v))
			}
		case reflect.Float32, reflect.Float64:
			if v, ok := m[key].(float64); ok {
				value.SetFloat(v)
			}
		case reflect.String:
			if v, ok := m[key].(string); ok {
				value.SetString(v)
			}
		case reflect.Slice:
			tmpSli := reflect.MakeSlice(sf.Type, 0, 0)
			tmpVal := reflect.New(sf.Type.Elem()).Interface()

			if _, ok := m[key].([]bson.M); !ok {
				break
			}

			for _, v := range m[key].([]bson.M) {
				err := Bson2struct(v, tmpVal)
				if err != nil {
					fmt.Println(err)
					break
				}
				tmpSli = reflect.Append(tmpSli, reflect.ValueOf(tmpVal).Elem())
			}
			rv.Field(i).Set(tmpSli)
		}
	}
	return

}

func Struct2bson(s interface{}) (m bson.M, err error) {
	rt := reflect.TypeOf(s)
	rv := reflect.ValueOf(s)

	for rt.Kind() == reflect.Ptr {
		rt = rt.Elem()
		rv = rv.Elem()
	}

	if rt.Kind() != reflect.Struct {
		err = fmt.Errorf("not struct %v", rt.Kind())
		return
	}

	m = make(bson.M)

	for i := 0; i < rt.NumField(); i++ {
		sf := rt.Field(i)
		k := sf.Tag.Get("json")
		k = strings.ToLower(k)
		kk := strings.Split(k, ",")
		if len(kk) < 1 {
			continue
		}
		key := kk[0]
		if key == "-" {
			continue
		}
		value := rv.Field(i)

		switch sf.Type.Kind() {
		case reflect.Bool:
			m[key] = value.Bool()
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			m[key] = value.Int()
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			if int64(value.Uint()) < 0 {
				panic(fmt.Sprintf("[%s] BSON has no uint64 type, and value is too large to fit correctly in an int64 ", sf.Name))
			}
			m[key] = int64(value.Uint())
		case reflect.Float32, reflect.Float64:
			m[key] = value.Float()
		case reflect.String:
			m[key] = value.String()
		case reflect.Slice:
			var sm []bson.M
			for i := 0; i < value.Len(); i++ {
				mi, err := Struct2bson(value.Index(i).Interface())
				if err != nil {
					fmt.Println(err)
					break
				}
				sm = append(sm, mi)
			}
			m[key] = sm
		default:

		}
	}

	return
}
