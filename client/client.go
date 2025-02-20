package client

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
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
	ReceiveFileAndSave(*bufio.Reader) error
	ValidateSavePath() error
}

type Client struct {
	ip       string
	port     int
	savepath string
	conn     net.Conn
	filemap  map[string]int64
}

func NewClient(ip string, port int, savepath string) *Client {
	return &Client{
		ip:       ip,
		port:     port,
		savepath: filepath.Clean(savepath),
		filemap:  make(map[string]int64, 16),
	}
}

func (c *Client) Start() error {
	conn, connErr := net.Dial("tcp", fmt.Sprintf("%s:%d", c.ip, c.port))
	if connErr != nil {
		return connErr
	}
	defer conn.Close()
	c.conn = conn

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
		fmt.Fprintf(os.Stderr, "Failed to receive file list: %s\n", recvErr)
		return
	}
	if !c.ShowFileList() || !c.PromptConfirmation() {
		return
	}
	if confirmErr := c.SendConfirmation(writer); confirmErr != nil {
		fmt.Fprintf(os.Stderr, "Failed to send confirmation: %s\n", confirmErr)
		return
	}
	for len(c.filemap) > 0 {
		if recvErr := c.ReceiveFileAndSave(reader); recvErr != nil {
			if errors.Is(recvErr, io.EOF) {
				return
			}
			fmt.Fprintf(os.Stderr, "Failed to receive file: %s\n", recvErr)
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
		fmt.Fprintf(os.Stderr, "No files to download\n")
		return false
	}

	totalSize := int64(0)
	for name, size := range c.filemap {
		fmt.Fprintf(os.Stdout, "[File] %s [Size] %s\n", name, utils.ToReadableSize(size))
		totalSize += size
	}
	fmt.Fprintf(os.Stdout, "[Total] %d files, %s\n", len(c.filemap), utils.ToReadableSize(totalSize))

	return true
}

func (c *Client) PromptConfirmation() bool {
	fmt.Fprintf(os.Stdout, "Confirm accept[Y/n]: ")
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
	confirmPkt := protocol.NewPacket(protocol.Confirm, "", 0)
	if _, confirmErr := confirmPkt.EnableAndWrite(writer); confirmErr != nil {
		return confirmErr
	}
	return writer.Flush()
}

func (c *Client) ReceiveFileAndSave(reader *bufio.Reader) error {
	metaPkt := new(protocol.Packet)
	if _, metaErr := metaPkt.ReadAndDecode(reader); metaErr != nil {
		return metaErr
	}
	filesize, ok := c.filemap[metaPkt.FileName]
	if !ok {
		return fmt.Errorf("file not found: %s", metaPkt.FileName)
	}

	filePath := filepath.Join(c.savepath, metaPkt.FileName)
	file, createErr := os.Create(filePath)
	if createErr != nil {
		return fmt.Errorf("[%s] cannot create file: %s", metaPkt.FileName, createErr)
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

func (c *Client) ValidateSavePath() error {
	savepathStat, statErr := os.Stat(c.savepath)
	if os.IsNotExist(statErr) || !savepathStat.IsDir() {
		return fmt.Errorf("[%s] path does not exist or is not a folder", c.savepath)
	}

	file, tmpErr := os.CreateTemp(c.savepath, ".write_test")
	if tmpErr != nil {
		return fmt.Errorf("[%s] path is not writable", c.savepath)
	}
	file.Close()
	if removeErr := os.Remove(file.Name()); removeErr != nil {
		return fmt.Errorf("[%s] path is not writable", c.savepath)
	}

	return nil
}
