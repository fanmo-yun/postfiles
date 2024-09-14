package datainfo

import (
	"encoding/json"
	"fmt"
	"os"
	"postfiles/exitcodes"
)

type DataInfo struct {
	Name string `json:"name"`
	Size int64  `json:"size"`
	Type int8   `json:"type"`
}

func NewInfo(name string, size int64, datatype int8) *DataInfo {
	return &DataInfo{name, size, datatype}
}

func EncodeJSON(info *DataInfo) []byte {
	jsonData, encodeErr := json.Marshal(info)
	if encodeErr != nil {
		fmt.Fprintf(os.Stderr, "Error encoding JSON: %s", encodeErr)
		os.Exit(exitcodes.ErrJsonEncoding)
	}
	return jsonData
}

func DecodeJSON(info []byte) *DataInfo {
	var fileinfo DataInfo
	if err := json.Unmarshal(info, &fileinfo); err != nil {
		fmt.Fprintf(os.Stderr, "Error unmarshal JSON: %s", err)
		os.Exit(exitcodes.ErrJsonUnmarshal)
	}
	return &fileinfo
}
