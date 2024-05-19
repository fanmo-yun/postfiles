package client

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"postfiles/api"

	"github.com/sirupsen/logrus"
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
		logrus.Fatal(connErr)
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
			logrus.Fatal(readErr)
		}

		switch msgType {
		case 0:
			jsonData, readErr := reader.ReadBytes('\n')
			if readErr != nil {
				if readErr == io.EOF {
					break
				}
				logrus.Fatal(readErr)
			}
			info = api.DecodeJSON(jsonData[:])
		case 1:
			if info == nil {
				logrus.Fatal("FileInfo not initialized")
			}
			fp, createErr := os.Create(filepath.Join(savepath, info.FileName))
			if createErr != nil {
				logrus.Fatal(createErr)
			}
			defer fp.Close()
			if _, copyErr := io.CopyN(fp, reader, info.FileSize); copyErr != nil {
				if copyErr == io.EOF {
					continue
				}
				logrus.Fatal(copyErr)
			}
		}
	}
}
