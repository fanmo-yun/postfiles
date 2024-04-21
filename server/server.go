package server

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"postfiles/api"
)

type Server struct {
	IP   string
	Port int
}

func NewServer(IP string, Port int) *Server {
	return &Server{IP, Port}
}

func (server Server) ServerRun(files []string) {
	listener, listenErr := net.Listen("tcp", fmt.Sprintf("%s:%d", server.IP, server.Port))
	if listenErr != nil {
		log.Fatal(listenErr)
	}

	for {
		conn, connErr := listener.Accept()
		if connErr != nil {
			log.Panic(connErr)
			continue
		}
		go server.serverhandler(conn, &files)
	}
}

func (server Server) serverhandler(conn net.Conn, fileList *[]string) {
	defer conn.Close()

	writer := bufio.NewWriter(conn)

	for _, value := range *fileList {
		filename, filesize := api.FileStat(value)

		if writeErr := writer.WriteByte(0); writeErr != nil {
			if writeErr == io.EOF {
				break
			}
			log.Fatal(writeErr)
		}

		if flushErr := writer.Flush(); flushErr != nil {
			log.Fatal(flushErr)
		}

		if _, writeErr := writer.Write(api.EncodeJSON(api.NewInfo(filename, filesize))); writeErr != nil {
			if writeErr == io.EOF {
				break
			}
			log.Fatal(writeErr)
		}

		if flushErr := writer.Flush(); flushErr != nil {
			log.Fatal(flushErr)
		}

		fp, openErr := os.Open(value)
		if openErr != nil {
			log.Fatal(openErr)
		}
		defer fp.Close()

		if writeErr := writer.WriteByte(1); writeErr != nil {
			if writeErr == io.EOF {
				break
			}
			log.Fatal(writeErr)
		}

		if flushErr := writer.Flush(); flushErr != nil {
			log.Fatal(flushErr)
		}

		limitedReader := &io.LimitedReader{R: fp, N: filesize}
		if _, copyErr := io.Copy(writer, limitedReader); copyErr != nil {
			if copyErr == io.EOF {
				break
			}
			log.Fatal(copyErr)
		}

		if flushErr := writer.Flush(); flushErr != nil {
			log.Fatal(flushErr)
		}
	}
}
