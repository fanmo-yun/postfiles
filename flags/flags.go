package flags

import (
	"flag"
	"fmt"
	"log"
	"os"
	"postfiles/api"
	"postfiles/server"
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
		fmt.Fprintf(os.Stderr, "Info: No Flag Parameters Are Treated as Files\n")
	}
	flag.Parse()
	args.Files = flag.Args()
}

func (args *Arguments) Handler() {
	if len(args.IP) == 0 {
		args.IP = api.GenIP()
	} else if !api.IsvalidIP(args.IP) {
		log.Fatalf("Ip incorrect: %s", args.IP)
	}

	if !api.IsvalidPort(args.Port) {
		log.Fatalf("Ip incorrect: %d", args.Port)
	}

	if strings.ToLower(args.Type) == "server" && len(args.Files) == 0 {
		log.Fatal("No Files")
	}

	if strings.ToLower(args.Type) == "client" && len(args.SavePath) == 0 {
		args.SavePath = api.GetDownloadPath()
	}
}

func (args *Arguments) Run() {
	switch strings.ToLower(args.Type) {
	case "server":
		fmt.Printf("server start: %s:%d\n", args.IP, args.Port)
		server := server.NewServer(args.IP, args.Port)
		server.ServerRun(args.Files)
	case "client":
		fmt.Println("client start")
	default:
		log.Fatal("unknown type")
	}
}
