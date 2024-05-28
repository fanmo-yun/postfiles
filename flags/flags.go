package flags

import (
	"flag"
	"fmt"
	"os"
	"postfiles/api"
	"postfiles/client"
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
		fmt.Fprintf(os.Stderr, "Invalid IP: %s\n", args.IP)
		os.Exit(1)
	}

	if !api.IsvalidPort(args.Port) {
		fmt.Fprintf(os.Stderr, "Invalid Port: %s\n", args.IP)
		os.Exit(1)
	}

	if strings.ToLower(args.Type) == "server" && len(args.Files) == 0 {
		fmt.Fprintf(os.Stderr, "No Files provided\n")
		os.Exit(1)
	}

	if strings.ToLower(args.Type) == "client" && len(args.SavePath) == 0 {
		args.SavePath = api.GetDownloadPath()
	}
}

func (args *Arguments) Run() {
	fmt.Fprintf(os.Stdout, "Application running on %s:%d ...\n", args.IP, args.Port)
	switch strings.ToLower(args.Type) {
	case "server":
		fmt.Fprintln(os.Stdout, "Starting in server mode")
		fmt.Printf("server start: %s:%d\n", args.IP, args.Port)
		server := server.NewServer(args.IP, args.Port)
		server.ServerRun(args.Files)
	case "client":
		fmt.Fprintln(os.Stdout, "Starting in client mode")
		fmt.Printf("client start: %s:%d\n", args.IP, args.Port)
		client := client.NewClient(args.IP, args.Port)
		client.ClientRun(args.SavePath)
	default:
		fmt.Fprintln(os.Stderr, "unknown type")
		os.Exit(1)
	}
}
