package protocol

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"os"
)

type DataInfo struct {
	Name string
	Size int64
	Type int8
}

func NewDataInfo(name string, size int64, datatype int8) *DataInfo {
	return &DataInfo{name, size, datatype}
}

func (data DataInfo) Encode() ([]byte, error) {
	bytes := new(bytes.Buffer)
	encodeErr := gob.NewEncoder(bytes).Encode(data)
	if encodeErr != nil {
		fmt.Fprintf(os.Stderr, "Error encode: %s", encodeErr)
		return nil, encodeErr
	}
	return bytes.Bytes(), nil
}

func (data *DataInfo) Decode(info []byte) error {
	decodeErr := gob.NewDecoder(bytes.NewReader(info)).Decode(data)
	if decodeErr != nil {
		fmt.Fprintf(os.Stderr, "Error decode: %s", decodeErr)
		return decodeErr
	}
	return nil
}
