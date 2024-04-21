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
			var jsonData [1024]byte
			size, readErr := reader.Read(jsonData[:])
			if readErr != nil {
				if readErr == io.EOF {
					break
				}
				log.Fatal("here")
			}
			info = api.DecodeJSON(jsonData[:size])
		case 1:
			if info == nil {
				log.Fatal("FileInfo not initialized")
			}
			limitedReader := &io.LimitedReader{R: reader, N: info.FileSize}
			fp, createErr := os.Create(filepath.Join(savepath, info.FileName))
			if createErr != nil {
				log.Fatal(createErr)
			}
			if _, copyErr := io.Copy(fp, limitedReader); copyErr != nil {
				if copyErr == io.EOF {
					continue
				}
				log.Fatal(copyErr)
			}
		}
	}
}
