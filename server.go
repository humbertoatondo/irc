// Copyright © 2016 Alan A. A. Donovan & Brian W. Kernighan.
// License: https://creativecommons.org/licenses/by-nc-sa/4.0/

// See page 254.
//!+

// Chat is a server that lets clients chat with each other.
package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"log"
	"net"
	"strings"
	"time"
)

var serverName = "irc-server >"

var usersList = make([]string, 0)
var channelsList = make(map[string]chan<- string)
var clientsIPList = make(map[string]string)

type client chan<- string // an outgoing message channel

var (
	entering = make(chan client)
	leaving  = make(chan client)
	messages = make(chan string) // all incoming client messages
)

func deleteUser(userName string) {
	for i, user := range usersList {
		if userName == user {
			usersList[i] = usersList[len(usersList)-1]
			usersList[len(usersList)-1] = ""
			usersList = usersList[:len(usersList)-1]
		}
	}

	//Remove channel from list
	_, channel := channelsList[userName]
	if channel {
		delete(channelsList, userName)
	}

	//Remove ip from list
	_, ip := clientsIPList[userName]
	if ip {
		delete(clientsIPList, userName)
	}
}

func getUserList(ch chan string) {
	for _, user := range usersList {
		if user != "" {
			ch <- serverName + " " + user
		}
	}
}

func isUserInList(username string) bool {
	for _, user := range usersList {
		if user == username {
			return true
		}
	}
	return false
}

//!+broadcaster
func broadcaster() {
	clients := make(map[client]bool) // all connected clients
	for {
		select {
		case msg := <-messages:
			// Broadcast incoming message to all
			// clients' outgoing message channels.
			for cli := range clients {
				cli <- msg
			}

		case cli := <-entering:
			clients[cli] = true

		case cli := <-leaving:
			delete(clients, cli)
			close(cli)
		}
	}
}

//!-broadcaster

//!+handleConn
func handleConn(conn net.Conn) {
	//clients := make(map[client]bool) // all connected clients
	ch := make(chan string) // outgoing client messages
	go clientWriter(conn, ch)

	who := ""
	usersList = append(usersList, "")
	fields := make([]string, 0)

	input := bufio.NewScanner(conn)
	for input.Scan() {
		if len(input.Text()) > 0 {
			fields = strings.Fields(input.Text())
		} else {
			fields = strings.Fields("-")
		}

		if usersList[len(usersList)-1] == "" {
			if !isUserInList(fields[0]) {
				usersList[len(usersList)-1] = fields[0]
				who = fields[0]

				channelsList[fields[0]] = ch
				clientsIPList[fields[0]] = conn.RemoteAddr().String()

				fmt.Println(serverName, "New connected user ["+fields[0]+"]")

				ch <- serverName + " Welcome to the Simple IRC Server" + who
				ch <- serverName + " Your user [" + who + "] is succesfully logged"
				messages <- serverName + " " + who + " has arrived"
				entering <- ch
			} else {
				ch <- serverName + " Username is already in use, please enter another username:"
			}
		} else if fields[0] == "/users" {
			getUserList(ch)
		} else if fields[0] == "/msg" && len(fields) > 2 {
			if isUserInList(fields[1]) {
				var buffer bytes.Buffer
				buffer.WriteString(who + ": ")
				for i := 2; i <= len(fields)-1; i++ {
					buffer.WriteString(fields[i] + " ")
				}
				channelsList[fields[1]] <- buffer.String()
			} else {
				ch <- serverName + " There is no user with that name."
			}
		} else if fields[0] == "/time" {
			t := time.Now()
			ch <- serverName + " " + t.Format("02/01/2006 15:04:05")
		} else if fields[0] == "/user" && len(fields) == 2 {
			ch <- serverName + " User: " + fields[1]
			ch <- serverName + " IP address: " + clientsIPList[fields[1]]
		} else {
			messages <- who + " > " + input.Text()
		}
	}
	// NOTE: ignoring potential errors from input.Err()

	leaving <- ch
	messages <- serverName + " " + who + " has left"
	fmt.Println(serverName, "["+fields[0]+"] left")
	deleteUser(who)
	conn.Close()
}

func clientWriter(conn net.Conn, ch <-chan string) {
	for msg := range ch {
		fmt.Fprintln(conn, msg) // NOTE: ignoring network errors
	}
}

//!-handleConn

//!+main
func main() {

	//Handle flags
	hostPtr := flag.String("host", "localhost:8000", "a string")
	portPtr := flag.String("port", "9000", "a string")
	flag.Parse()
	//End handling flags

	host := *hostPtr + ":" + *portPtr
	listener, err := net.Listen("tcp", host)
	if err != nil {
		log.Fatal(err)
	}

	go broadcaster()
	fmt.Println(serverName, "Simple IRC Server started at", host)
	fmt.Println(serverName, "Ready for receiving new clients")
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Print(err)
			continue
		}
		go handleConn(conn)
	}
}

// Requirements
/*
github.com/go-vgo/robotgo
https://github.com/go-vgo/robotgo#installation
*/
