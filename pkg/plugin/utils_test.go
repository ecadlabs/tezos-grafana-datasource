package plugin

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type T struct {
	Field0 string
	*TT
}

type TT struct {
	Field1 string `json:"field1"`
}

func TestPickFieldByName(t *testing.T) {
	values := []*T{
		{
			Field0: "f00",
			TT: &TT{
				Field1: "f10",
			},
		},
		{
			Field0: "f01",
			TT: &TT{
				Field1: "f11",
			},
		},
	}
	require.Equal(t, []string{"f00", "f01"}, pickFieldByName(values, "Field0"))
	require.Equal(t, []string{"f10", "f11"}, pickFieldByName(values, "field1"))
}
