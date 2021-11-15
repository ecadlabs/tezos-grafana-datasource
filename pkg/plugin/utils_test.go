package plugin

import (
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type Struct0 struct {
	Field0 string    `json:"field0"`
	Field1 time.Time `json:"field1"`
	Field2 *Struct1  `json:"field2"`
	*EmbeddedStruct
}

type Struct1 struct {
	Field0 int64 `json:"field0"`
}

type EmbeddedStruct struct {
	Field3 int64 `json:"field3"`
}

func TestFieldSelectors(t *testing.T) {
	f := getStructFields(&Struct0{})
	assert.Equal(t, []*structField{
		{
			Selector: []string{"field0"},
			Type:     reflect.TypeOf(""),
		},
		{
			Selector: []string{"field1"},
			Type:     reflect.TypeOf(time.Time{}),
		},
		{
			Selector: []string{"field2", "field0"},
			Type:     reflect.TypeOf(int64(0)),
		},
		{
			Selector: []string{"field3"},
			Type:     reflect.TypeOf(int64(0)),
		},
	}, f)
}
