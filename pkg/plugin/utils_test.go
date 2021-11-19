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
			selector: []string{"field0"},
			typ:      reflect.TypeOf(""),
		},
		{
			selector: []string{"field1"},
			typ:      reflect.TypeOf(time.Time{}),
		},
		{
			selector: []string{"field2", "field0"},
			typ:      reflect.TypeOf(int64(0)),
		},
		{
			selector: []string{"field3"},
			typ:      reflect.TypeOf(int64(0)),
		},
	}, f)
}
