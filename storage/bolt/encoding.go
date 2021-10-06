package bolt

import (
	"bytes"
	"encoding"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"math"
	"reflect"
)

type Codec interface {
	Marshal(v interface{}) ([]byte, error)
	Unmarshal(data []byte, v interface{}) error
}

type BinaryCodec struct{}

var be = binary.BigEndian

func isFixed(t reflect.Type) bool {
	k := t.Kind()
	return k >= reflect.Bool && k <= reflect.Complex128 || k == reflect.Array && isFixed(t.Elem())
}

func putFixed(buf []byte, v reflect.Value) {
	k := v.Kind()
	switch {
	case k >= reflect.Int && k <= reflect.Int64:
		i := v.Int()
		switch k {
		case reflect.Int8:
			buf[0] = byte(i)
		case reflect.Int16:
			be.PutUint16(buf, uint16(i))
		case reflect.Int32:
			be.PutUint32(buf, uint32(i))
		case reflect.Int, reflect.Int64:
			be.PutUint64(buf, uint64(i))
		}
	case k >= reflect.Uint && k <= reflect.Uintptr:
		i := v.Uint()
		switch k {
		case reflect.Uint8:
			buf[0] = byte(i)
		case reflect.Uint16:
			be.PutUint16(buf, uint16(i))
		case reflect.Uint32:
			be.PutUint32(buf, uint32(i))
		case reflect.Uint, reflect.Uint64, reflect.Uintptr:
			be.PutUint64(buf, i)
		}
	case k == reflect.Bool:
		if v.Bool() {
			buf[0] = 1
		}
	case k == reflect.Float32:
		be.PutUint32(buf, math.Float32bits(float32(v.Float())))
	case k == reflect.Float64:
		be.PutUint64(buf, math.Float64bits(v.Float()))
	case k == reflect.Complex64:
		be.PutUint32(buf, math.Float32bits(float32(real(v.Complex()))))
		be.PutUint32(buf[4:], math.Float32bits(float32(imag(v.Complex()))))
	case k == reflect.Complex128:
		be.PutUint64(buf, math.Float64bits(real(v.Complex())))
		be.PutUint64(buf[8:], math.Float64bits(imag(v.Complex())))
	case k == reflect.Array:
		sz := v.Type().Elem().Size()
		for i := 0; i < v.Len(); i++ {
			putFixed(buf[i*int(sz):], v.Index(i))
		}
	}
}

func getFixed(buf []byte, v reflect.Value) {
	k := v.Kind()
	switch {
	case k >= reflect.Int && k <= reflect.Int64:
		var i int64
		switch k {
		case reflect.Int8:
			i = int64(buf[0])
		case reflect.Int16:
			i = int64(be.Uint16(buf))
		case reflect.Int32:
			i = int64(be.Uint32(buf))
		case reflect.Int, reflect.Int64:
			i = int64(be.Uint64(buf))
		}
		v.SetInt(i)
	case k >= reflect.Uint && k <= reflect.Uintptr:
		var i uint64
		switch k {
		case reflect.Uint8:
			i = uint64(buf[0])
		case reflect.Uint16:
			i = uint64(be.Uint16(buf))
		case reflect.Uint32:
			i = uint64(be.Uint32(buf))
		case reflect.Uint, reflect.Uint64, reflect.Uintptr:
			i = uint64(be.Uint64(buf))
		}
		v.SetUint(i)
	case k == reflect.Bool:
		if buf[0] == 1 {
			v.SetBool(true)
		}
	case k == reflect.Float32:
		v.SetFloat(float64(math.Float32frombits(be.Uint32(buf))))
	case k == reflect.Float64:
		v.SetFloat(math.Float64frombits(be.Uint64(buf)))
	case k == reflect.Complex64:
		v.SetComplex(complex(float64(math.Float32frombits(be.Uint32(buf))), float64(math.Float32frombits(be.Uint32(buf[4:])))))
	case k == reflect.Complex128:
		v.SetComplex(complex(math.Float64frombits(be.Uint64(buf)), math.Float64frombits(be.Uint64(buf[8:]))))
	case k == reflect.Array:
		sz := v.Type().Elem().Size()
		for i := 0; i < v.Len(); i++ {
			getFixed(buf[i*int(sz):], v.Index(i))
		}
	}
}

func (BinaryCodec) Marshal(val interface{}) ([]byte, error) {
	if b, ok := val.(encoding.BinaryMarshaler); ok {
		return b.MarshalBinary()
	}
	v := reflect.Indirect(reflect.ValueOf(val))
	switch {
	case isFixed(v.Type()):
		buf := make([]byte, v.Type().Size())
		putFixed(buf, v)
		return buf, nil
	case v.Kind() == reflect.String:
		return []byte(v.String()), nil
	case v.Kind() == reflect.Slice:
		if v.Type().Elem().Kind() == reflect.Uint8 {
			return v.Bytes(), nil
		} else if isFixed(v.Type().Elem()) {
			sz := v.Type().Elem().Size()
			buf := make([]byte, int(sz)*v.Len())
			for i := 0; i < v.Len(); i++ {
				putFixed(buf[i*int(sz):], v.Index(i))
			}
			return buf, nil
		}
	}
	// fall back to Gob
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(val); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (BinaryCodec) Unmarshal(data []byte, val interface{}) error {
	v := reflect.ValueOf(val)
	if v.Kind() != reflect.Ptr {
		return fmt.Errorf("codec: not a pointer: %v", v.Type())
	}
	v = v.Elem()
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			v.Set(reflect.New(v.Type().Elem()))
		}
		v = v.Elem()
	}
	if b, ok := v.Addr().Interface().(encoding.BinaryUnmarshaler); ok {
		return b.UnmarshalBinary(data)
	}
	switch {
	case isFixed(v.Type()):
		if len(data) < int(v.Type().Size()) {
			return fmt.Errorf("buffer is too short: %d", len(data))
		}
		getFixed(data, v)
		return nil
	case v.Kind() == reflect.String:
		v.SetString(string(data))
		return nil
	case v.Kind() == reflect.Slice:
		if v.Type().Elem().Kind() == reflect.Uint8 {
			v.SetBytes(data)
			return nil
		} else if isFixed(v.Type().Elem()) {
			sz := v.Type().Elem().Size()
			if v.IsNil() {
				ln := len(data) / int(sz)
				v.Set(reflect.MakeSlice(v.Type(), ln, ln))
			} else if len(data) < int(sz)*v.Len() {
				return fmt.Errorf("buffer is too short: %d", len(data))
			}
			for i := 0; i < v.Len(); i++ {
				getFixed(data[i*int(sz):], v.Index(i))
			}
			return nil
		}
	}
	// fall back to Gob
	return gob.NewDecoder(bytes.NewReader(data)).Decode(val)
}

type GobCodec struct{}

func (GobCodec) Marshal(val interface{}) ([]byte, error) {
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(val); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (GobCodec) Unmarshal(data []byte, val interface{}) error {
	return gob.NewDecoder(bytes.NewReader(data)).Decode(val)
}
