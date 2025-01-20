package client

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"postfiles/datainfo"
	"postfiles/exitcodes"
	"postfiles/utils"
	"strings"
)

type Client struct {
	IP    string
	Port  int
	count int16
	size  int64
}

func NewClient(IP string, Port int) *Client {
	return &Client{IP, Port, 0, 0}
}

func (c Client) ClientRun(savepath string) {
	conn, connErr := net.Dial("tcp", fmt.Sprintf("%s:%d", c.IP, c.Port))
	if connErr != nil {
		fmt.Fprintf(os.Stderr, "Failed to connect: %s\n", connErr)
		os.Exit(exitcodes.ErrClient)
	}
	defer conn.Close()

	c.clientHandle(conn, savepath)
}

func (c Client) clientHandle(conn net.Conn, savepath string) {
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
			os.Exit(exitcodes.ErrClient)
		}
		decMsg := new(datainfo.DataInfo)
		decodeErr := decMsg.Decode(msg)
		if decodeErr != nil {
			fmt.Fprintf(os.Stderr, "Failed to decode message: %v\n", decodeErr)
			continue
		}

		switch decMsg.Type {
		case datainfo.File_Info_Data:
			c.receiveFileData(reader, savepath, decMsg)

		case datainfo.File_Count:
			c.handleFileCount(decMsg)

		default:
			fmt.Fprintf(os.Stdout, "All file count: %d, All file size: %.2f MB\n\n", c.count, utils.ToMB(c.size))
			if !c.handleConfirm() {
				break LOOP
			}
			if err := c.sendConfirm(writer); err != nil {
				break LOOP
			}
		}
	}
}

func (c *Client) receiveFileData(reader *bufio.Reader, savepath string, info *datainfo.DataInfo) {
	filePath := filepath.Join(savepath, info.Name)
	fp, createErr := os.Create(filePath)
	if createErr != nil {
		fmt.Fprintf(os.Stderr, "Failed to create file %s: %s\n", filePath, createErr)
		os.Exit(exitcodes.ErrClient)
	}
	defer fp.Close()

	bar := utils.CreateBar(info.Size, info.Name)

	if _, copyErr := io.CopyN(io.MultiWriter(fp, bar), reader, info.Size); copyErr != nil {
		if copyErr != io.EOF {
			fmt.Fprintf(os.Stderr, "Failed to copy file data for %s: %s\n", filePath, copyErr)
			os.Exit(exitcodes.ErrClient)
		}
	}
}

func (c *Client) handleFileCount(info *datainfo.DataInfo) {
	c.count += 1
	c.size += info.Size
	fmt.Fprintf(os.Stdout, "[%d] - %s - %.2f MB\n", c.count, info.Name, utils.ToMB(info.Size))
}

func (c *Client) handleConfirm() bool {
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

func (c *Client) sendConfirm(w *bufio.Writer) error {
	confirmInfo := datainfo.NewDataInfo("Confirm_Accept", 0, datainfo.Confirm_Accept)
	encodedInfo, encodeErr := confirmInfo.Encode()
	if encodeErr != nil {
		return encodeErr
	}

	if _, writeErr := w.Write(encodedInfo); writeErr != nil {
		return fmt.Errorf("failed to write file info: %s", writeErr)
	}
	if writeErr := w.WriteByte('\n'); writeErr != nil {
		return fmt.Errorf("failed to write newline after file info: %s", writeErr)
	}
	if flushErr := w.Flush(); flushErr != nil {
		return fmt.Errorf("failed to flush writer after file info: %s", flushErr)
	}
	return nil
}
