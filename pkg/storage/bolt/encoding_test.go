package bolt

import (
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncoderDecoder(t *testing.T) {
	var values = []interface{}{
		true,
		int(123),
		int8(123),
		int16(-12345),
		int32(123456),
		int64(-1234567),
		uint(123),
		uint8(123),
		uint16(12345),
		uint32(123456),
		uint64(1234567),
		uintptr(12345678),
		float32(1.2345),
		float64(1.2345678),
		complex64(1.2345 + 2.3456i),
		complex128(1.2345678 + 2.3456789i),
		[]byte("hello"),
		string("hello"),
		[4]int{1, 2, 3, 4},
		struct {
			X int
			Y int
		}{X: 1, Y: 2},
		time.Unix(819163440, 0),
	}
	t.Parallel()
	t.Run("Direct", func(t *testing.T) {
		for _, value := range values {
			t.Run(reflect.TypeOf(value).String(), func(t *testing.T) {
				var codec BinaryCodec
				b, err := codec.Marshal(value)
				require.NoError(t, err)
				result := reflect.New(reflect.TypeOf(value))
				require.NoError(t, codec.Unmarshal(b, result.Interface()))
				assert.Equal(t, value, result.Elem().Interface())
			})
		}
	})
	t.Run("Indirect", func(t *testing.T) {
		for _, value := range values {
			t.Run(reflect.TypeOf(value).String(), func(t *testing.T) {
				var codec BinaryCodec
				b, err := codec.Marshal(value)
				require.NoError(t, err)
				result := reflect.New(reflect.PtrTo(reflect.TypeOf(value)))
				require.NoError(t, codec.Unmarshal(b, result.Interface()))
				assert.Equal(t, value, result.Elem().Elem().Interface())
			})
		}
	})
}
