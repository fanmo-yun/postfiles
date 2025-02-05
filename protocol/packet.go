package protocol

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/json"
	"io"
)

type PacketInterface interface {
	Encode() ([]byte, uint32, error)
	Decode([]byte) error
	EnableAndWrite(*bufio.Writer) error
	ReadAndDecode(*bufio.Reader) error
}

type Packet struct {
	DataType DataType `json:"DataType"`
	FileName string   `json:"FileName"`
	FileSize int64    `json:"FileSize"`
}

func NewPacket(DataType DataType, FileName string, FileSize int64) *Packet {
	return &Packet{
		DataType: DataType,
		FileName: FileName,
		FileSize: FileSize,
	}
}

func (dp *Packet) Encode() ([]byte, uint32, error) {
	bytes := new(bytes.Buffer)
	if err := json.NewEncoder(bytes).Encode(dp); err != nil {
		return nil, 0, err
	}
	return bytes.Bytes(), uint32(bytes.Len()), nil
}

func (dp *Packet) Decode(Bytes []byte) error {
	if err := json.NewDecoder(bytes.NewReader(Bytes)).Decode(dp); err != nil {
		return err
	}
	return nil
}

func (dp *Packet) EnableAndWrite(writer *bufio.Writer) error {
	encodedPacket, packetLen, encodeErr := dp.Encode()
	if encodeErr != nil {
		return encodeErr
	}
	if binWriteErr := binary.Write(writer, binary.LittleEndian, packetLen); binWriteErr != nil {
		return binWriteErr
	}
	if _, writeErr := writer.Write(encodedPacket); writeErr != nil {
		return writeErr
	}
	return nil
}

func (dp *Packet) ReadAndDecode(reader *bufio.Reader) error {
	var decLength uint32
	if readErr := binary.Read(reader, binary.LittleEndian, &decLength); readErr != nil {
		return readErr
	}
	decData := make([]byte, decLength)
	_, readErr := io.ReadFull(reader, decData)
	if readErr != nil {
		return readErr
	}
	return dp.Decode(decData)
}
