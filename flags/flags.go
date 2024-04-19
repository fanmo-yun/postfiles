package flags

import (
	"flag"
	"fmt"
	"log"
	"os"
	"postfiles/api"
	"strings"
)

type Arguments struct {
	Type     string
	IP       string
	Port     int
	Files    []string
	SavePath string
}

func Newflags() *Arguments {
	return &Arguments{}
}

func (args *Arguments) Parser() {
	flag.StringVar(&args.Type, "type", "Server", "Server Or Client")
	flag.StringVar(&args.IP, "ip", "", "IP Address (default \"Ip currently in use\")")
	flag.IntVar(&args.Port, "port", 8877, "Port Number")
	flag.StringVar(&args.SavePath, "save", "", "Save Path, Only valid for clients (default \"System Download Path\")")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()
	args.Files = flag.Args()
}

func (args *Arguments) Handler() {
	if args.IP == "" {
		args.IP = api.GenIP()
	} else if !api.IsvalidIP(args.IP) {
		log.Fatalf("Ip incorrect: %s", args.IP)
	}

	if !api.IsvalidPort(args.Port) {
		log.Fatalf("Ip incorrect: %d", args.Port)
	}
}

func (args *Arguments) Run() {
	if strings.ToLower(args.Type) == "server" {
		// s := server.NewServer(args.IP, args.Port)
		// s.ServerRun()
		fmt.Printf("Server Start: %s:%d\n", args.IP, args.Port)
	}
}
