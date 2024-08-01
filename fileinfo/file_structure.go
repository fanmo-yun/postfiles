package fileinfo

import (
	"encoding/json"
	"fmt"
	"os"
	"postfiles/exitcodes"
)

type FileInfo struct {
	FileName string `json:"filename"`
	FileSize int64  `json:"filesize"`
}

func NewInfo(filename string, filesize int64) *FileInfo {
	return &FileInfo{filename, filesize}
}

func EncodeJSON(info *FileInfo) []byte {
	jsonData, encodeErr := json.Marshal(info)
	if encodeErr != nil {
		fmt.Fprintf(os.Stderr, "Error encoding JSON: %s", encodeErr)
		os.Exit(exitcodes.ErrJsonEncoding)
	}
	return jsonData
}

func DecodeJSON(info []byte) *FileInfo {
	var fileinfo FileInfo
	if err := json.Unmarshal(info, &fileinfo); err != nil {
		fmt.Fprintf(os.Stderr, "Error unmarshal JSON: %s", err)
		os.Exit(exitcodes.ErrJsonUnmarshal)
	}
	return &fileinfo
}
