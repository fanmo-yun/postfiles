package protocol

import (
	"bufio"
	"encoding/binary"
	"encoding/json"
	"io"
)

type PacketInterface interface {
	encode() ([]byte, uint32, error)
	decode([]byte) error
	EncodeAndWrite(*bufio.Writer) error
	ReadAndDecode(*bufio.Reader) error
	Is(DataType) bool
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

func (p *Packet) encode() ([]byte, uint32, error) {
	packet, err := json.Marshal(p)
	if err != nil {
		return nil, 0, err
	}
	return packet, uint32(len(packet)), nil
}

func (p *Packet) decode(Bytes []byte) error {
	return json.Unmarshal(Bytes, p)
}

func (p *Packet) EncodeAndWrite(writer *bufio.Writer) (int, error) {
	encPkt, pktLen, encodeErr := p.encode()
	if encodeErr != nil {
		return 0, encodeErr
	}
	if binWriteErr := binary.Write(writer, binary.LittleEndian, pktLen); binWriteErr != nil {
		return 0, binWriteErr
	}
	n, writeErr := writer.Write(encPkt)
	if writeErr != nil {
		return n, writeErr
	}
	return n, writer.Flush()
}

func (p *Packet) ReadAndDecode(reader *bufio.Reader) (int, error) {
	var pktLen uint32
	if readErr := binary.Read(reader, binary.LittleEndian, &pktLen); readErr != nil {
		return 0, readErr
	}
	decBuf := make([]byte, pktLen)
	n, readErr := io.ReadFull(reader, decBuf)
	if readErr != nil {
		return 0, readErr
	}
	return n, p.decode(decBuf)
}

func (p *Packet) Is(dt DataType) bool {
	return p.DataType == dt
}
