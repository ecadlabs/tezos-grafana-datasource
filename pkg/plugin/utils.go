package plugin

import (
	"encoding"
	"encoding/json"
	"math/big"
	"reflect"
	"strings"
	"time"
)

type structField struct {
	Selector []string
	Type     reflect.Type
}

var (
	timeType      = reflect.TypeOf(time.Time{})
	bigIntType    = reflect.TypeOf((*big.Int)(nil)).Elem()
	bigFloatType  = reflect.TypeOf((*big.Float)(nil)).Elem()
	bigRatType    = reflect.TypeOf((*big.Rat)(nil)).Elem()
	textMarshaler = reflect.TypeOf((*encoding.TextMarshaler)(nil)).Elem()
	jsonMarshaler = reflect.TypeOf((*json.Marshaler)(nil)).Elem()
)

// collect fields with corresponding selectors which are convertable to CUE types and then to Grafana types
func getStructTypeFields(t reflect.Type) (fields []*structField) {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	fields = make([]*structField, 0, t.NumField())
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if !field.IsExported() {
			continue
		}
		ft := field.Type
		if ft.Kind() == reflect.Ptr {
			ft = ft.Elem()
		}
		name := field.Name
		if jn := strings.Split(field.Tag.Get("json"), ",")[0]; jn != "" {
			name = jn
		}
		if k := ft.Kind(); k >= reflect.Bool && k <= reflect.Float64 || k == reflect.String ||
			ft == timeType || ft == bigIntType || ft == bigFloatType || ft == bigRatType ||
			ft.Implements(jsonMarshaler) || ft.Implements(textMarshaler) {
			fields = append(fields, &structField{Selector: []string{name}, Type: ft})
		} else if ft.Kind() == reflect.Struct {
			nestedFields := getStructTypeFields(ft)
			if field.Anonymous {
				fields = append(fields, nestedFields...)
			} else {
				for _, f := range nestedFields {
					fields = append(fields, &structField{
						Type:     f.Type,
						Selector: append([]string{name}, f.Selector...),
					})
				}
			}
		}
	}
	return
}

func getStructFields(v interface{}) (fields []*structField) {
	return getStructTypeFields(reflect.TypeOf(v))
}
