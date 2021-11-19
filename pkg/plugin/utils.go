package plugin

import (
	"encoding"
	"reflect"

	"github.com/ecadlabs/jtree"
)

type structField struct {
	selector []string
	typ      reflect.Type
}

var (
	textMarshaler = reflect.TypeOf((*encoding.TextMarshaler)(nil)).Elem()
)

// collect fields with corresponding selectors which are convertable to CUE types and then to Grafana types
func getStructTypeFields(t reflect.Type) []*structField {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	out := make([]*structField, 0)
	for _, field := range jtree.VisibleFields(t) {
		ft := field.Type
		for ft.Kind() == reflect.Ptr {
			ft = ft.Elem()
		}
		k := ft.Kind()
		if k >= reflect.Bool && k <= reflect.Float64 || k == reflect.String || ft.Implements(textMarshaler) || reflect.PtrTo(ft).Implements(textMarshaler) {
			out = append(out, &structField{selector: []string{field.Name}, typ: ft})
		} else if ft.Kind() == reflect.Struct {
			for _, f := range getStructTypeFields(ft) {
				out = append(out, &structField{
					typ:      f.typ,
					selector: append([]string{field.Name}, f.selector...),
				})
			}
		}
	}
	return out
}

func getStructFields(v interface{}) (fields []*structField) {
	return getStructTypeFields(reflect.TypeOf(v))
}
