package model

import (
	"bytes"
	"encoding/binary"
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

func (val *Int64) UnmarshalBinary(data []byte) error {
	n, err := binary.ReadVarint(bytes.NewReader(data))
	if err != nil {
		return err
	}
	*val = Int64(n)
	return nil
}

func (val Int64) MarshalBinary() (data []byte, err error) {
	buf := make([]byte, binary.MaxVarintLen64)
	n := binary.PutVarint(buf, int64(val))
	return buf[:n], nil
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

type Base58 []byte

func (b Base58) String() string {
	return EncodeBase58Check(b)
}

func (val *Base58) UnmarshalText(text []byte) (err error) {
	*val, err = DecodeBase58Check(string(text))
	return
}

func (val Base58) MarshalText() ([]byte, error) {
	return []byte(EncodeBase58Check(val)), nil
}
