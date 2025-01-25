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
	encodErr := gob.NewEncoder(bytes).Encode(data)
	if encodErr != nil {
		fmt.Fprintf(os.Stderr, "Error encode: %s", encodErr)
		return nil, encodErr
	}
	return bytes.Bytes(), nil
}

func (data *DataInfo) Decode(info []byte) error {
	decodErr := gob.NewDecoder(bytes.NewReader(info)).Decode(data)
	if decodErr != nil {
		fmt.Fprintf(os.Stderr, "Error decode: %s", decodErr)
		return decodErr
	}
	return nil
}
