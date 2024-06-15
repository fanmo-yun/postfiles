package client

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"postfiles/fileinfo"
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
		fmt.Fprintf(os.Stderr, "Failed to connect: %v\n", connErr)
		os.Exit(2)
	}
	defer conn.Close()

	reader := bufio.NewReader(conn)
	var (
		info     *fileinfo.FileInfo
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
		case fileinfo.File_Info:
			jsonData, readErr := reader.ReadBytes('\n')
			if readErr != nil {
				if readErr == io.EOF {
					break
				}
				fmt.Fprintf(os.Stderr, "Failed to read JSON data: %v\n", readErr)
				os.Exit(2)
			}
			info = fileinfo.DecodeJSON(jsonData[:])

		case fileinfo.File_Data:
			if info == nil {
				fmt.Fprintf(os.Stderr, "FileInfo not initialized\n")
				os.Exit(2)
			}
			fp, createErr := os.Create(filepath.Join(savepath, info.FileName))
			if createErr != nil {
				fmt.Fprintf(os.Stderr, "Failed to create file: %v\n", createErr)
				os.Exit(2)
			}

			bar := init_bar(info.FileSize, info.FileName)

			if _, copyErr := io.CopyN(io.MultiWriter(fp, bar), reader, info.FileSize); copyErr != nil {
				if copyErr == io.EOF {
					continue
				}
				fmt.Fprintf(os.Stderr, "Failed to copy file data: %v\n", copyErr)
				os.Exit(2)
			}

			fp.Close()

		case fileinfo.File_Count:
			for {
				jsonData, readErr := reader.ReadBytes('\n')
				if readErr != nil {
					if readErr == io.EOF {
						break
					}
					fmt.Fprintf(os.Stderr, "Failed to read JSON data: %v\n", readErr)
					os.Exit(2)
				}
				info = fileinfo.DecodeJSON(jsonData[:])
				if info.FileSize != -1 {
					allCount += 1
					allSize += info.FileSize
				} else {
					break
				}
			}

			fmt.Fprintf(os.Stdout, "All file count: %d, All file size: %d\nconfirm recv[Y/n]: ", allCount, allSize)
			readin()
		}
	}
}

func readin() string {
	reader := bufio.NewReader(os.Stdin)

	data, readErr := reader.ReadString('\n')
	if readErr != nil {
		fmt.Fprintln(os.Stderr, "Error reading input:", readErr)
		os.Exit(1)
	}

	return strings.TrimRight(data, "\r\n")
}
