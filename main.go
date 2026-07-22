package main

import (
	"fmt"
	"log"
	"net"
	"os"

	"github.com/govim/govim"
)

func main() {
	f, err := os.OpenFile("/tmp/rover.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}

	l, err := net.Listen("tcp", "localhost:8765")
	if err != nil {
		log.Fatal(err)
	}
	conn, err := l.Accept()
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	// set to ioutil.Discard if not in debug mode

	p := &VimMdPlugin{}
	g, err := govim.NewGovim(p, conn, conn, f, &p.tomb)
	if err != nil {
		fmt.Printf("failed to create rover.vim instance: %v", err)
	}

	g.Run()
}
