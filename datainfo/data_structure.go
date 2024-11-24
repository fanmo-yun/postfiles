package datainfo

import (
	"encoding/json"
	"fmt"
	"os"
)

type DataInfo struct {
	Name string `json:"name"`
	Size int64  `json:"size"`
	Type int8   `json:"type"`
}

func NewInfo(name string, size int64, datatype int8) *DataInfo {
	return &DataInfo{name, size, datatype}
}

func EncodeJSON(info *DataInfo) ([]byte, error) {
	jsonData, encodeErr := json.Marshal(info)
	if encodeErr != nil {
		fmt.Fprintf(os.Stderr, "Error encoding JSON: %s", encodeErr)
		return nil, encodeErr
	}
	return jsonData, nil
}

func DecodeJSON(info []byte) (*DataInfo, error) {
	var fileinfo DataInfo
	if unmarshalErr := json.Unmarshal(info, &fileinfo); unmarshalErr != nil {
		fmt.Fprintf(os.Stderr, "Error unmarshal JSON: %s", unmarshalErr)
		return nil, unmarshalErr
	}
	return &fileinfo, nil
}
