package client

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"postfiles/log"
	"postfiles/protocol"
	"postfiles/utils"
	"strings"
)

type ClientInterface interface {
	Start() error
	HandleTransfer()
	ReceiveFileList(*bufio.Reader) error
	ShowFileList() bool
	PromptConfirmation() bool
	SendConfirmation(*bufio.Writer) error
	ReceiveFileAndSave(*bufio.Reader, *bufio.Writer) error
	ValidateWritable() error
}

type Client struct {
	ip       string
	port     int
	savepath *os.Root
	conn     net.Conn
	filemap  map[string]int64
}

func NewClient(ip string, port int, savepath *os.Root) *Client {
	return &Client{
		ip:       ip,
		port:     port,
		savepath: savepath,
		filemap:  make(map[string]int64, 16),
	}
}

func (c *Client) Start() error {
	address := net.JoinHostPort(c.ip, fmt.Sprintf("%d", c.port))
	conn, connErr := net.Dial("tcp", address)
	if connErr != nil {
		return connErr
	}
	defer conn.Close()
	c.conn = conn
	log.PrintToOut("Client start at %s, Save Path: %s\n", address, c.savepath.Name())

	c.HandleTransfer()
	return nil
}

func (c *Client) HandleTransfer() {
	reader := bufio.NewReaderSize(c.conn, 32*1024)
	writer := bufio.NewWriterSize(c.conn, 32*1024)

	if recvErr := c.ReceiveFileList(reader); recvErr != nil {
		if errors.Is(recvErr, io.EOF) {
			return
		}
		log.PrintToErr("Failed to receive file list: %s\n", recvErr)
		return
	}
	if !c.ShowFileList() || !c.PromptConfirmation() {
		return
	}
	if confirmErr := c.SendConfirmation(writer); confirmErr != nil {
		log.PrintToErr("Failed to send confirmation: %s\n", confirmErr)
		return
	}
	for len(c.filemap) > 0 {
		if recvErr := c.ReceiveFileAndSave(reader, writer); recvErr != nil {
			if errors.Is(recvErr, io.EOF) {
				return
			} else if errors.Is(recvErr, os.ErrExist) {
				continue
			}
			log.PrintToErr("Failed to receive file: %s\n", recvErr)
			return
		}
	}
}

func (c *Client) ReceiveFileList(reader *bufio.Reader) error {
	for {
		recvPkt := new(protocol.Packet)
		if _, readErr := recvPkt.ReadAndDecode(reader); readErr != nil {
			return readErr
		}

		switch recvPkt.DataType {
		case protocol.FileQuantity:
			c.filemap[recvPkt.FileName] = recvPkt.FileSize

		case protocol.EndOfTransmission:
			return nil

		default:
			return errors.New("unknown message type")
		}
	}
}

func (c *Client) ShowFileList() bool {
	if len(c.filemap) == 0 {
		log.PrintToErr("No files to download\n")
		return false
	}

	totalSize := int64(0)
	for name, size := range c.filemap {
		log.PrintToOut("[File] %s [Size] %s\n", name, utils.ToReadableSize(size))
		totalSize += size
	}
	log.PrintToOut("[Total] %d files, %s\n", len(c.filemap), utils.ToReadableSize(totalSize))

	return true
}

func (c *Client) PromptConfirmation() bool {
	log.PrintToOut("Confirm accept[Y/n]: ")
	confirm, readinErr := utils.Readin()
	if readinErr != nil {
		return false
	}

	switch strings.ToLower(confirm) {
	case "y", "yes":
		return true
	case "n", "no":
		return false
	default:
		return false
	}
}

func (c *Client) SendConfirmation(writer *bufio.Writer) error {
	confirmPkt := protocol.NewPacket(protocol.ConfirmAccept, "", 0)
	_, confirmErr := confirmPkt.EncodeAndWrite(writer)
	return confirmErr
}

func (c *Client) ReceiveFileAndSave(reader *bufio.Reader, writer *bufio.Writer) error {
	metaPkt := new(protocol.Packet)
	if _, metaErr := metaPkt.ReadAndDecode(reader); metaErr != nil {
		return metaErr
	}
	filesize, ok := c.filemap[metaPkt.FileName]
	if !ok {
		return fmt.Errorf("file not found in client: %s", metaPkt.FileName)
	}

	if fstat, err := c.savepath.Stat(metaPkt.FileName); err == nil && !fstat.IsDir() {
		procStr, procErr := utils.ProcessString(metaPkt.FileName)
		if procErr != nil {
			return procErr
		}
		log.PrintToOut("--skip-- [%s] <-- %s\n", procStr, os.ErrExist)
		rejPkt := protocol.NewPacket(protocol.RejectFile, "", 0)
		if _, writeErr := rejPkt.EncodeAndWrite(writer); writeErr != nil {
			return writeErr
		}
		return os.ErrExist
	}

	accPkt := protocol.NewPacket(protocol.AcceptFile, "", 0)
	if _, writeErr := accPkt.EncodeAndWrite(writer); writeErr != nil {
		return writeErr
	}

	file, createErr := c.savepath.Create(metaPkt.FileName)
	if createErr != nil {
		return fmt.Errorf("[%s] cannot be create: %s", metaPkt.FileName, createErr)
	}
	defer file.Close()

	pb, pbErr := utils.CreateProcessBar(filesize, metaPkt.FileName)
	if pbErr != nil {
		return pbErr
	}
	mw := io.MultiWriter(file, pb)
	if _, copyErr := io.CopyN(mw, reader, filesize); copyErr != nil {
		if errors.Is(copyErr, io.EOF) {
			return io.EOF
		}
		return fmt.Errorf("[%s] failed to copy file data: %s", metaPkt.FileName, copyErr)
	}
	delete(c.filemap, metaPkt.FileName)
	return nil
}

func (c *Client) ValidateWritable() error {
	file, tmpErr := os.CreateTemp(c.savepath.Name(), "write.tmp")
	if tmpErr != nil {
		return fmt.Errorf("[%s] path is not writable", c.savepath.Name())
	}
	file.Close()
	if removeErr := os.Remove(file.Name()); removeErr != nil {
		return fmt.Errorf("[%s] path is not writable", c.savepath.Name())
	}
	return nil
}
