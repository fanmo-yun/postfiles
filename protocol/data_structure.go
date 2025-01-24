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
	infobytes := new(bytes.Buffer)
	Encoderr := gob.NewEncoder(infobytes).Encode(data)
	if Encoderr != nil {
		fmt.Fprintf(os.Stderr, "Error encode: %s", Encoderr)
		return nil, Encoderr
	}
	return infobytes.Bytes(), nil
}

func (data *DataInfo) Decode(info []byte) error {
	Decoderr := gob.NewDecoder(bytes.NewReader(info)).Decode(data)
	if Decoderr != nil {
		fmt.Fprintf(os.Stderr, "Error decode: %s", Decoderr)
		return Decoderr
	}
	return nil
}
