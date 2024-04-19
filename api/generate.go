package api

import (
	"log"
	"net"
	"strings"
)

func GenIP() string {
	conn, connErr := net.Dial("udp", "114.114.114.114:80")
	if connErr != nil {
		log.Fatal(connErr)
	}
	defer conn.Close()
	return strings.Split(conn.LocalAddr().String(), ":")[0]
}
