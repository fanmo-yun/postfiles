package client

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"postfiles/api"
)

type Client struct {
	IP   string
	Port int
}

func NewClient(IP string, Port int) *Client {
	return &Client{IP, Port}
}

func (client Client) ClientRun(savepath string) {
	conn, connErr := net.Dial("tcp", fmt.Sprintf("%s:%d", client.IP, client.Port))
	if connErr != nil {
		log.Fatal(connErr)
	}
	defer conn.Close()

	reader := bufio.NewReader(conn)
	var info *api.FileInfo

	for {
		msgType, readErr := reader.ReadByte()
		if readErr != nil {
			if readErr == io.EOF {
				break
			}
			log.Fatal(readErr)
		}

		switch msgType {
		case 0:
			jsonData, readErr := reader.ReadBytes('\n')
			if readErr != nil {
				if readErr == io.EOF {
					break
				}
				log.Fatal("here")
			}
			info = api.DecodeJSON(jsonData[:])
		case 1:
			if info == nil {
				log.Fatal("FileInfo not initialized")
			}
			fp, createErr := os.Create(filepath.Join(savepath, info.FileName))
			if createErr != nil {
				log.Fatal(createErr)
			}
			defer fp.Close()
			if _, copyErr := io.CopyN(fp, reader, info.FileSize); copyErr != nil {
				if copyErr == io.EOF {
					continue
				}
				log.Fatal(copyErr)
			}
		}
	}
}
