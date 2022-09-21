package main

import (
	"flag"
	"fmt"
	"go/chat/client"
	"go/chat/server"
	"os"
	"strings"
)

const (
	IP   = "127.0.0.1"
	PORT = "3569"
)

func usage() {
	fmt.Print("Unrecognized option \n\n")
	fmt.Println("Usage : go run main.go --mode <mode>")
	fmt.Print("- mode : \"server\" or \"client \n\n")
	fmt.Println("example to run a server: go run main.go --mode server")
	fmt.Println("example to run a client: go run main.go --mode client")
}

func main() {
	var mode string
	flag.StringVar(&mode, "mode", "client", "--mode client or --mode server")
	flag.Parse()
	if strings.ToLower(mode) == "server" {
		server := server.New(IP, PORT)
		server.Run()
	} else if strings.ToLower(mode) == "client" {
		client := client.New(IP, PORT)
		client.Connect()
		client.CreateUsername()
	} else {
		usage()
		os.Exit(2)
	}
}
