package protocol

type DataInfo struct {
	Name string `json:"name"`
	Size int64  `json:"size"`
	Type int8   `json:"type"`
}

func NewDataInfo(name string, size int64, datatype int8) *DataInfo {
	return &DataInfo{name, size, datatype}
}

func (v DataInfo) Encode() ([]byte, error) {
	jsonBytes, marshalErr := v.MarshalJSON()
	if marshalErr != nil {
		return nil, marshalErr
	}
	return jsonBytes, nil
}

func (v *DataInfo) Decode(jsonBytes []byte) error {
	v.UnmarshalJSON(jsonBytes)
	return nil
}
