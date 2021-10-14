package plugin

import (
	"fmt"
	"reflect"
	"strings"
)

func getStructField(t reflect.Type, fieldName string) *reflect.StructField {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	for _, f := range reflect.VisibleFields(t) {
		if !f.IsExported() {
			continue
		}
		name := f.Name
		if jn := strings.Split(f.Tag.Get("json"), ",")[0]; jn != "" {
			name = jn
		}
		if name == fieldName {
			return &f
		}
	}
	return nil
}

func pickFieldByName(src interface{}, fieldName string) interface{} {
	val := reflect.Indirect(reflect.ValueOf(src))
	if val.Kind() != reflect.Slice {
		panic(fmt.Sprintf("slice expected: %v", val.Type()))
	}
	elem := val.Type().Elem()
	if elem.Kind() == reflect.Ptr {
		elem = elem.Elem()
	}
	if elem.Kind() != reflect.Struct {
		panic(fmt.Sprintf("slice of struct expected: %v", val.Type()))
	}
	field := getStructField(elem, fieldName)
	if field == nil {
		return nil
	}
	res := reflect.MakeSlice(reflect.SliceOf(field.Type), val.Len(), val.Len())
	for i := 0; i < val.Len(); i++ {
		v := reflect.Indirect(val.Index(i))
		res.Index(i).Set(v.FieldByIndex(field.Index))
	}
	return res.Interface()
}
