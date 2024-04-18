package main

import "postfiles/server"

func main() {
	s := server.NewServer("127.0.0.1", 3000)
	s.ServerRun()
}
