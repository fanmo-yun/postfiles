package client

import (
	"bufio"
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
	ClientRun()
	HandleConnection(conn net.Conn)
	ReceiveFile(reader *bufio.Reader, info *protocol.DataInfo)
	ProcessFileCount(info *protocol.DataInfo)
	PromptConfirm() bool
	SendConfirmation(writer *bufio.Writer) error
}

type Client struct {
	IP       string
	Port     int
	SavePath string
	Count    int16
	Size     int64
}

func NewClient(IP string, Port int, SavePath string) *Client {
	return &Client{IP, Port, SavePath, 0, 0}
}

func (c *Client) ClientRun() {
	conn, connErr := net.Dial("tcp", fmt.Sprintf("%s:%d", c.IP, c.Port))
	if connErr != nil {
		fmt.Fprintf(os.Stderr, "Failed to connect: %s\n", connErr)
		os.Exit(utils.ErrClient)
	}
	defer conn.Close()

	c.HandleConnection(conn)
}

func (c *Client) HandleConnection(conn net.Conn) {
	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)

LOOP:
	for {
		msg, readErr := reader.ReadBytes('\n')
		if readErr != nil {
			if readErr == io.EOF {
				break LOOP
			}
			fmt.Fprintf(os.Stderr, "Failed to read message type: %s\n", readErr)
			os.Exit(utils.ErrClient)
		}
		decMsg := new(protocol.DataInfo)
		decodeErr := decMsg.Decode(msg)
		if decodeErr != nil {
			fmt.Fprintf(os.Stderr, "Failed to decode message: %s\n", decodeErr)
			continue
		}

		switch decMsg.Type {
		case protocol.File_Info_Data:
			c.ReceiveFile(reader, decMsg)

		case protocol.File_Count:
			c.ProcessFileCount(decMsg)

		default:
			fmt.Fprintf(os.Stdout, "All file count: %d, All file size: %s\n\n", c.Count, utils.ToReadableSize(c.Size))
			if c.Count == 0 || !c.PromptConfirm() {
				break LOOP
			}
			if err := c.SendConfirmation(writer); err != nil {
				break LOOP
			}
		}
	}
}

func (c *Client) ReceiveFile(reader *bufio.Reader, info *protocol.DataInfo) {
	filePath := filepath.Join(c.SavePath, info.Name)
	fp, createErr := os.Create(filePath)
	if createErr != nil {
		fmt.Fprintf(os.Stderr, "Failed to create file %s: %s\n", filePath, createErr)
		os.Exit(utils.ErrClient)
	}
	defer fp.Close()

	mw := io.MultiWriter(fp, utils.CreateProcessBar(info.Size, info.Name))

	if _, copyErr := io.CopyN(mw, reader, info.Size); copyErr != nil {
		if copyErr != io.EOF {
			fmt.Fprintf(os.Stderr, "Failed to copy file data for %s: %s\n", filePath, copyErr)
			os.Exit(utils.ErrClient)
		}
	}
}

func (c *Client) ProcessFileCount(info *protocol.DataInfo) {
	c.Count += 1
	c.Size += info.Size
	fmt.Fprintf(os.Stdout, "[%d] - %s - %s\n", c.Count, info.Name, utils.ToReadableSize(info.Size))
}

func (c *Client) PromptConfirm() bool {
	fmt.Fprintf(os.Stdout, "Confirm accept[Y/n]: ")
	confirm := utils.Readin()
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
	confirmInfo := protocol.NewDataInfo("Confirm_Accept", 0, protocol.Confirm_Accept)
	encodedInfo, encodeErr := confirmInfo.Encode()
	if encodeErr != nil {
		return encodeErr
	}

	if _, writeErr := writer.Write(encodedInfo); writeErr != nil {
		return fmt.Errorf("failed to write file info: %s", writeErr)
	}
	if writeErr := writer.WriteByte('\n'); writeErr != nil {
		return fmt.Errorf("failed to write newline after file info: %s", writeErr)
	}
	if flushErr := writer.Flush(); flushErr != nil {
		return fmt.Errorf("failed to flush writer after file info: %s", flushErr)
	}
	return nil
}
