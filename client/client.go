package client

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"postfiles/exitcodes"
	"postfiles/fileinfo"
	"postfiles/utils"
	"strings"
)

type Client struct {
	IP   string
	Port int
}

func NewClient(IP string, Port int) *Client {
	return &Client{IP, Port}
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
		msgType, readErr := reader.ReadByte()
		if readErr != nil {
			if readErr == io.EOF {
				break
			}
			fmt.Fprintf(os.Stderr, "Failed to read message type: %s\n", readErr)
			os.Exit(exitcodes.ErrClient)
		}

		switch msgType {
		case fileinfo.File_Info_Data:
			info := c.readFileInfo(reader)
			c.receiveFileData(reader, savepath, info)

		case fileinfo.File_Count:
			c.handleFileCount(reader)
			if !c.handleConfirm() {
				break LOOP
			}
			if err := c.sendConfirm(writer); err != nil {
				break LOOP
			}
		}
	}
}

func (c *Client) readFileInfo(reader *bufio.Reader) *fileinfo.FileInfo {
	jsonData, readErr := reader.ReadBytes('\n')
	if readErr != nil {
		if readErr == io.EOF {
			os.Exit(exitcodes.ErrClient)
		}
		fmt.Fprintf(os.Stderr, "Failed to read JSON data: %s\n", readErr)
		os.Exit(exitcodes.ErrClient)
	}
	return fileinfo.DecodeJSON(jsonData[:])
}

func (c *Client) receiveFileData(reader *bufio.Reader, savepath string, info *fileinfo.FileInfo) {
	filePath := filepath.Join(savepath, info.FileName)
	fp, createErr := os.Create(filePath)
	if createErr != nil {
		fmt.Fprintf(os.Stderr, "Failed to create file %s: %s\n", filePath, createErr)
		os.Exit(exitcodes.ErrClient)
	}
	defer fp.Close()

	bar := utils.CreateBar(info.FileSize, info.FileName)

	if _, copyErr := io.CopyN(io.MultiWriter(fp, bar), reader, info.FileSize); copyErr != nil {
		if copyErr != io.EOF {
			fmt.Fprintf(os.Stderr, "Failed to copy file data for %s: %s\n", filePath, copyErr)
			os.Exit(exitcodes.ErrClient)
		}
	}
}

func (c *Client) handleFileCount(reader *bufio.Reader) {
	count := uint16(0)
	size := int64(0)

	for {
		jsonData, readErr := reader.ReadBytes('\n')
		if readErr != nil {
			if readErr == io.EOF {
				os.Exit(exitcodes.ErrClient)
			}
			fmt.Fprintf(os.Stderr, "Failed to read JSON data: %s\n", readErr)
			os.Exit(exitcodes.ErrClient)
		}
		info := fileinfo.DecodeJSON(jsonData[:])
		if info.FileSize != fileinfo.End_Of_Transmission {
			count += 1
			size += info.FileSize
			fmt.Fprintf(os.Stdout, "[%d] - %s - %.2f MB\n", count, info.FileName, utils.ToMB(info.FileSize))
		} else {
			break
		}
	}

	fmt.Fprintf(os.Stdout, "All file count: %d, All file size: %.2f MB\n\n", count, utils.ToMB(size))
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
	confirmInfo := fileinfo.NewInfo("CONFIRM_ACCEPT", fileinfo.Confirm_Accept)
	encodedInfo := fileinfo.EncodeJSON(confirmInfo)

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
