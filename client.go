// Copyright © 2016 Alan A. A. Donovan & Brian W. Kernighan.
// License: https://creativecommons.org/licenses/by-nc-sa/4.0/

// See page 227.

// Netcat is a simple read/write client for TCP servers.
package main

import (
	"flag"
	"io"
	"log"
	"net"
	"os"

	"github.com/go-vgo/robotgo"
)

var clear map[string]func() //create a map for storing clear funcs

//!+
func main() {

	//Handle flags
	userPtr := flag.String("user", "foo", "a string")
	serverPtr := flag.String("server", "localhost:9000", "a string")

	flag.Parse()
	//End handling flags

	conn, err := net.Dial("tcp", *serverPtr)
	if err != nil {
		log.Fatal(err)
	}

	done := make(chan struct{})

	robotgo.TypeStr(*userPtr)
	robotgo.KeyTap("enter")
	print("\033[H\033[2J")

	go func() {
		io.Copy(os.Stdout, conn) // NOTE: ignoring errors
		log.Println("done")
		done <- struct{}{} // signal the main goroutine
	}()

	mustCopy(conn, os.Stdin)
	conn.Close()
	<-done // wait for background goroutine to finish
}

//!-

func mustCopy(dst io.Writer, src io.Reader) {
	if _, err := io.Copy(dst, src); err != nil {
		log.Fatal(err)
	}
}
