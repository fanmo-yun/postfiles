package client

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"postfiles/fileinfo"
	"postfiles/utils"

	"github.com/schollz/progressbar/v3"
)

type Client struct {
	IP   string
	Port int
}

func NewClient(IP string, Port int) *Client {
	return &Client{IP, Port}
}

func (client Client) ClientRun(savepath string) {
	w := utils.GetBarWidth() - 40

	conn, connErr := net.Dial("tcp", fmt.Sprintf("%s:%d", client.IP, client.Port))
	if connErr != nil {
		fmt.Fprintf(os.Stderr, "Failed to connect: %v\n", connErr)
		os.Exit(1)
	}
	defer conn.Close()

	reader := bufio.NewReader(conn)
	var info *fileinfo.FileInfo

	for {
		msgType, readErr := reader.ReadByte()
		if readErr != nil {
			if readErr == io.EOF {
				break
			}
			fmt.Fprintf(os.Stderr, "Failed to read message type: %v\n", readErr)
			os.Exit(1)
		}

		switch msgType {
		case fileinfo.File_Info:
			jsonData, readErr := reader.ReadBytes('\n')
			if readErr != nil {
				if readErr == io.EOF {
					break
				}
				fmt.Fprintf(os.Stderr, "Failed to read JSON data: %v\n", readErr)
				os.Exit(1)
			}
			info = fileinfo.DecodeJSON(jsonData[:])
		case fileinfo.File_Data:
			if info == nil {
				fmt.Fprintf(os.Stderr, "FileInfo not initialized\n")
				os.Exit(1)
			}
			fp, createErr := os.Create(filepath.Join(savepath, info.FileName))
			if createErr != nil {
				fmt.Fprintf(os.Stderr, "Failed to create file: %v\n", createErr)
				os.Exit(1)
			}
			defer fp.Close()

			bar := progressbar.NewOptions64(
				info.FileSize,
				progressbar.OptionSetDescription("Receiving file: "+utils.TruncateString(info.FileName, w-20)),
				progressbar.OptionSetWidth(w),
				progressbar.OptionShowBytes(true),
				progressbar.OptionSetPredictTime(true),
				progressbar.OptionShowCount(),
				progressbar.OptionOnCompletion(func() {
					fmt.Println()
				}),
				progressbar.OptionSetRenderBlankState(true),
			)

			if _, copyErr := io.CopyN(io.MultiWriter(fp, bar), reader, info.FileSize); copyErr != nil {
				if copyErr == io.EOF {
					continue
				}
				fmt.Fprintf(os.Stderr, "Failed to copy file data: %v\n", copyErr)
				os.Exit(1)
			}
		}
	}
}
