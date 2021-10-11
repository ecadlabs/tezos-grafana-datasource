package model

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var tests = []struct {
	vec string
	s   string
}{
	{
		vec: "00eb15231dfceb60925886b67d065299925915aeb172c06647",
		s:   "1NS17iag9jJgTHD1VXjvLCEnZuQ3rJDE9L",
	},
	{
		vec: "00000000000000000000",
		s:   "1111111111",
	},
	{
		vec: "000111d38e5fc9071ffcd20b4a763cc9ae4f252bb4e48fd66a835e252ada93ff480d6dd43dc62a641155a5",
		s:   "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz",
	},
}

var checkTests = []struct {
	vec string
	s   string
}{
	{
		vec: "025a7991a6caedf5419d01100e4587f0d4d9fc84b4749a",
		s:   "KT1MruMYHugk6x7qWQGeFKoV4fuarhTfoV6t",
	},
}

func TestEncodeBase58(t *testing.T) {
	for _, tt := range tests {
		v, _ := hex.DecodeString(tt.vec)
		assert.Equal(t, tt.s, EncodeBase58(v))
	}
}

func TestDecodeBase58(t *testing.T) {
	for _, tt := range tests {
		v, _ := hex.DecodeString(tt.vec)
		d, err := DecodeBase58(tt.s)
		require.NoError(t, err)
		assert.Equal(t, v, d)
	}
}

func TestEncodeBase58Check(t *testing.T) {
	for _, tt := range checkTests {
		v, _ := hex.DecodeString(tt.vec)
		assert.Equal(t, tt.s, EncodeBase58Check(v))
	}
}

func TestDecodeBase58Check(t *testing.T) {
	for _, tt := range checkTests {
		v, _ := hex.DecodeString(tt.vec)
		d, err := DecodeBase58Check(tt.s)
		require.NoError(t, err)
		assert.Equal(t, v, d)
	}
}
