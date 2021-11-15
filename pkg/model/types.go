package model

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
