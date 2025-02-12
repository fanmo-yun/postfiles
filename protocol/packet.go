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
	if encodeErr := json.NewEncoder(bytes).Encode(dp); encodeErr != nil {
		return nil, 0, encodeErr
	}
	return bytes.Bytes(), uint32(bytes.Len()), nil
}

func (dp *Packet) Decode(Bytes []byte) error {
	return json.NewDecoder(bytes.NewReader(Bytes)).Decode(dp)
}

func (dp *Packet) EnableAndWrite(writer *bufio.Writer) (int, error) {
	encPkt, pktLen, encodeErr := dp.Encode()
	if encodeErr != nil {
		return 0, encodeErr
	}
	if binWriteErr := binary.Write(writer, binary.LittleEndian, pktLen); binWriteErr != nil {
		return 0, binWriteErr
	}
	return writer.Write(encPkt)
}

func (dp *Packet) ReadAndDecode(reader *bufio.Reader) (int, error) {
	var pktLen uint32
	if readErr := binary.Read(reader, binary.LittleEndian, &pktLen); readErr != nil {
		return 0, readErr
	}
	decBuf := make([]byte, pktLen)
	n, readErr := io.ReadFull(reader, decBuf)
	if readErr != nil {
		return 0, readErr
	}
	return n, dp.Decode(decBuf)
}
