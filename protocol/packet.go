package protocol

import (
	"bufio"
	"encoding/binary"
	"encoding/json"
	"io"
)

type Packet struct {
	DataType DataType `json:"type"`
	FileName string   `json:"file,omitempty"`
	FileSize int64    `json:"size,omitempty"`
}

func NewPacket(t DataType, name string, size int64) *Packet {
	return &Packet{DataType: t, FileName: name, FileSize: size}
}

func (p *Packet) encode() ([]byte, error) {
	return json.Marshal(p)
}

func (p *Packet) decode(Bytes []byte) error {
	return json.Unmarshal(Bytes, p)
}

func (p *Packet) EncodeAndWrite(w *bufio.Writer) error {
	body, err := p.encode()
	if err != nil {
		return err
	}
	if err := binary.Write(w, binary.LittleEndian, uint32(len(body))); err != nil {
		return err
	}
	_, err = w.Write(body)
	return err
}

func (p *Packet) ReadAndDecode(r *bufio.Reader) error {
	var pktLen uint32
	if readErr := binary.Read(r, binary.LittleEndian, &pktLen); readErr != nil {
		return readErr
	}
	decBuf := make([]byte, pktLen)
	_, readErr := io.ReadFull(r, decBuf)
	if readErr != nil {
		return readErr
	}
	return p.decode(decBuf)
}

func (p *Packet) TypeIs(dt DataType) bool {
	return p.DataType == dt
}
