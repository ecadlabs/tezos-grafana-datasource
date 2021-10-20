package plugin

// a glue between CUE and Grafana types

import (
	"fmt"
	"time"

	"cuelang.org/go/cue"
	"github.com/grafana/grafana-plugin-sdk-go/data"
)

type fieldConverter interface {
	Set(i int, val cue.Value) error
	Field() *data.Field
}

type boolField data.Field

func (f *boolField) Set(i int, val cue.Value) error {
	v, err := val.Bool()
	if err != nil {
		return err
	}
	(*data.Field)(f).Set(i, v)
	return nil
}

func (f *boolField) Field() *data.Field { return (*data.Field)(f) }

type intField data.Field

func (f *intField) Set(i int, val cue.Value) error {
	v, err := val.Int64()
	if err != nil {
		return err
	}
	(*data.Field)(f).Set(i, v)
	return nil
}

func (f *intField) Field() *data.Field { return (*data.Field)(f) }

type floatField data.Field

func (f *floatField) Set(i int, val cue.Value) error {
	v, err := val.Float64()
	if err != nil {
		return err
	}
	(*data.Field)(f).Set(i, v)
	return nil
}

func (f *floatField) Field() *data.Field { return (*data.Field)(f) }

type stringField data.Field

func (f *stringField) Set(i int, val cue.Value) error {
	v, err := val.String()
	if err != nil {
		return err
	}
	(*data.Field)(f).Set(i, v)
	return nil
}

func (f *stringField) Field() *data.Field { return (*data.Field)(f) }

type timeField data.Field

func (f *timeField) Set(i int, val cue.Value) error {
	v, err := val.String()
	if err != nil {
		return err
	}
	var t time.Time
	if err := t.UnmarshalText([]byte(v)); err != nil {
		return err
	}
	(*data.Field)(f).Set(i, t)
	return nil
}

func (f *timeField) Field() *data.Field { return (*data.Field)(f) }

func newFieldFromFieldType(name string, p data.FieldType, n int) *data.Field {
	f := data.NewFieldFromFieldType(p, n)
	f.Name = name
	return f
}

func newFieldConverter(name string, val cue.Value, size int) (fieldConverter, error) {
	if val.Kind() == cue.StringKind {
		// guess type by syntax
		var t time.Time
		v, _ := val.String()
		if err := t.UnmarshalText([]byte(v)); err == nil {
			return (*timeField)(newFieldFromFieldType(name, data.FieldTypeTime, size)), nil
		} else {
			return (*stringField)(newFieldFromFieldType(name, data.FieldTypeString, size)), nil
		}
	}
	switch val.Kind() {
	case cue.BoolKind:
		return (*boolField)(newFieldFromFieldType(name, data.FieldTypeBool, size)), nil
	case cue.IntKind:
		return (*intField)(newFieldFromFieldType(name, data.FieldTypeInt64, size)), nil
	case cue.FloatKind:
		return (*floatField)(newFieldFromFieldType(name, data.FieldTypeFloat64, size)), nil
	}
	return nil, fmt.Errorf("unsupported type: %v", val.Kind())
}
