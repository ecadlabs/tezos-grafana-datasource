package storage

import (
	"encoding/hex"
	"encoding/json"
	"math/big"
	"strconv"
)

type Int64 int64

func (val *Int64) UnmarshalText(text []byte) error {
	v, err := strconv.ParseInt(string(text), 10, 64)
	if err != nil {
		return err
	}
	*val = Int64(v)
	return nil
}

func (val Int64) MarshalText() ([]byte, error) {
	return []byte(strconv.FormatInt(int64(val), 10)), nil
}

type BigInt struct {
	big.Int
}

func (b *BigInt) UnmarshalJSON(text []byte) error {
	if string(text) == "null" {
		return nil
	}
	var tmp string
	if err := json.Unmarshal(text, &tmp); err != nil {
		return err
	}
	return b.Int.UnmarshalText([]byte(tmp))
}

func (b *BigInt) MarshalJSON() ([]byte, error) {
	text, err := b.Int.MarshalText()
	if err != nil {
		return nil, err
	}
	return json.Marshal(string(text))
}

type Bytes []byte

func (val *Bytes) UnmarshalText(text []byte) (err error) {
	*val, err = hex.DecodeString(string(text))
	return
}

func (val Bytes) MarshalText() ([]byte, error) {
	return []byte(hex.EncodeToString(val)), nil
}
