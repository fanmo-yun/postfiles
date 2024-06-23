package client

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"postfiles/fileinfo"
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
		fmt.Fprintf(os.Stderr, "Failed to connect: %v\n", connErr)
		os.Exit(2)
	}
	defer conn.Close()

	c.clientHandle(conn, savepath)
}

func (c Client) clientHandle(conn net.Conn, savepath string) {
	reader := bufio.NewReader(conn)
	var (
		allCount int   = 0
		allSize  int64 = 0
	)

	for {
		msgType, readErr := reader.ReadByte()
		if readErr != nil {
			if readErr == io.EOF {
				break
			}
			fmt.Fprintf(os.Stderr, "Failed to read message type: %v\n", readErr)
			os.Exit(2)
		}

		switch msgType {
		case fileinfo.File_Info_Data:
			info := c.readFileInfo(reader)
			c.receiveFileData(reader, savepath, info)

		case fileinfo.File_Count:
			c.handleFileCount(reader, &allCount, &allSize)
		}
	}
}

func (c *Client) readFileInfo(reader *bufio.Reader) *fileinfo.FileInfo {
	jsonData, readErr := reader.ReadBytes('\n')
	if readErr != nil {
		if readErr == io.EOF {
			os.Exit(2)
		}
		fmt.Fprintf(os.Stderr, "Failed to read JSON data: %v\n", readErr)
		os.Exit(2)
	}
	return fileinfo.DecodeJSON(jsonData[:])
}

func (c *Client) receiveFileData(reader *bufio.Reader, savepath string, info *fileinfo.FileInfo) {
	filePath := filepath.Join(savepath, info.FileName)
	fp, createErr := os.Create(filePath)
	if createErr != nil {
		fmt.Fprintf(os.Stderr, "Failed to create file %s: %v\n", filePath, createErr)
		os.Exit(2)
	}
	defer fp.Close()

	bar := init_bar(info.FileSize, info.FileName)

	if _, copyErr := io.CopyN(io.MultiWriter(fp, bar), reader, info.FileSize); copyErr != nil {
		if copyErr != io.EOF {
			fmt.Fprintf(os.Stderr, "Failed to copy file data for %s: %v\n", filePath, copyErr)
			os.Exit(2)
		}
	}
}

func (c *Client) handleFileCount(reader *bufio.Reader, allCount *int, allSize *int64) {
	for {
		jsonData, readErr := reader.ReadBytes('\n')
		if readErr != nil {
			if readErr == io.EOF {
				break
			}
			fmt.Fprintf(os.Stderr, "Failed to read JSON data: %v\n", readErr)
			os.Exit(2)
		}
		info := fileinfo.DecodeJSON(jsonData[:])
		if info.FileSize != -1 {
			*allCount += 1
			*allSize += info.FileSize
			fmt.Fprintf(os.Stdout, "[%d] - %s - %d\n", allCount, info.FileName, info.FileSize)
		} else {
			break
		}
	}

	fmt.Fprintf(os.Stdout, "All file count: %d, All file size: %d\n", *allCount, *allSize)
}
