package api

import (
	"encoding/json"
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
		// logrus.Fatal(encodeErr)
	}
	return jsonData
}

func DecodeJSON(info []byte) *FileInfo {
	var fileinfo FileInfo
	if err := json.Unmarshal(info, &fileinfo); err != nil {
		// logrus.Fatal(err)
	}
	return &fileinfo
}
