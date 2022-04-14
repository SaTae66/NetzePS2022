package util

import (
	"bytes"
	"encoding/binary"
)

func ReadToStruct[T any](r *bytes.Reader) (T, error) {
	var x T
	err := binary.Read(r, binary.LittleEndian, &x)
	return x, err
}
