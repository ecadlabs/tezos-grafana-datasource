package client

import (
	"encoding/hex"
	"encoding/json"
	"math/big"
	"strconv"
)

type Int64Array []int64

func (val *Int64Array) UnmarshalJSON(text []byte) error {
	if string(text) == "null" {
		return nil
	}
	var tmp []string
	if err := json.Unmarshal(text, &tmp); err != nil {
		return err
	}
	*val = make(Int64Array, len(tmp))
	for i, s := range tmp {
		var err error
		if (*val)[i], err = strconv.ParseInt(s, 10, 64); err != nil {
			return err
		}
	}
	return nil
}

func (val Int64Array) MarshalJSON() ([]byte, error) {
	tmp := make([]string, len(val))
	for i, v := range val {
		tmp[i] = strconv.FormatInt(v, 10)
	}
	return json.Marshal(tmp)
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
