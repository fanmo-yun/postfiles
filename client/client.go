package client

import (
	"log/slog"
	"net"
	"os"
)

type Client struct {
	address string
	saveDir *os.Root
	fileMap map[string]int64
}

func NewClient(address string, saveDir *os.Root) *Client {
	return &Client{
		address: address,
		saveDir: saveDir,
		fileMap: make(map[string]int64, 16),
	}
}

func (c *Client) Start() error {
	defer c.saveDir.Close()
	conn, connErr := net.Dial("tcp", c.address)
	if connErr != nil {
		return connErr
	}
	defer conn.Close()
	slog.Info("Client start", "address", c.address, "save_dir", c.saveDir.Name())

	c.handleTransfer(conn)
	return nil
}
