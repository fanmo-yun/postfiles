package client

import (
	"fmt"
	"log"
	"net"
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
}
