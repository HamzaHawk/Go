package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
)

//Client represents a client
type Client struct {
	conn     net.Conn
	nickname string
	ch       chan string
}

//Slave represents a client
type Slave struct {
	conn     net.Conn
	nickname string
	ch       chan string
}

func main() {
	ln, err := net.Listen("tcp", ":3000")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	//msgchan is used for all the messages any user types
	//the message is first populated by handleConnection and passed on to handleMessages
	msgchan := make(chan string)
	msgchanslave := make(chan string)

	//A channel to keep track of Client connections, clients are added to this channel
	//handleMessages then iterates through clients and for each appends to its channel
	//the messages sent by other users, broadcast
	addchanclient := make(chan Client)
	removechanclient := make(chan Client)
	addchanslave := make(chan Slave)
	removechanslave := make(chan Slave)

	go handleMessagesClient(msgchan,msgchanslave, addchanclient, removechanclient, addchanslave, removechanslave)
	//go handleMessagesSlave(msgchan, addchanslave, removechanslave)

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println(err)
			continue
		}
		//for each client, we do have a separate handleConnection goroutine
		go handleConnection(conn, msgchan,msgchanslave, addchanclient,addchanslave, removechanclient,removechanslave)
	}
}

//ReadLinesInto is a method on Client type
//it keeps waiting for user to input a line, ch chan is the msgchannel
//it formats and writes the message to the channel
func (c Client) ReadLinesInto(ch chan<- string,removechanclient chan<- Client) {
	bufc := bufio.NewReader(c.conn)
	for {
		line, err := bufc.ReadString('\n')
		if err != nil {
			fmt.Println("some error reading")
			removechanclient <- c
			break
		}
		//ch <- fmt.Sprintf("%s: %s", c.nickname, line)
		log.Printf("Client says: %s", line)
		ch <- fmt.Sprintf("%s", line)
	}
}

//WriteLinesFrom is a method
//each client routine is writing to channel
func (c Client) WriteLinesFrom(ch <-chan string, removechanclient chan<- Client) {
	for msg := range ch {
		_, err := io.WriteString(c.conn, msg)
		if err != nil {
			removechanclient <- c
			return
		}
	}
}


//////////////////////////////////////////////////////////////////////////////
//ReadLinesInto is a method on Client type
//it keeps waiting for user to input a line, ch chan is the msgchannel
//it formats and writes the message to the channel
func (c Slave) ReadLinesInto(ch chan<- string) {
	bufc := bufio.NewReader(c.conn)
	for {
		line, err := bufc.ReadString('\n')
		if err != nil {
			break
		}
		ch <- fmt.Sprintf("%s: %s", c.nickname, line)
	}
}

//WriteLinesFrom is a method
//each client routine is writing to channel
func (c Slave) WriteLinesFrom(ch <-chan string, removechanslave chan<- Slave) {
	for msg := range ch {
		_, err := io.WriteString(c.conn, msg)
		if err != nil {
			removechanslave <- c
			return
		}
	}
}
///////////////////////////////////////////////////////////////////////////////


func promptNick(c net.Conn, bufc *bufio.Reader) string {
	//io.WriteString(c, "\033[1;30;41mWelcome to the fancy demo chat!\033[0m\n")
	//io.WriteString(c, "What is your nick? ")
	nick, _, _ := bufc.ReadLine()
	return string(nick)
}

//the core one
func handleConnection(c net.Conn, msgchan chan<- string, msgchanslave chan<- string, addchanclient chan<- Client, addchanslave chan<- Slave, removechanclient chan<- Client,removechanslave chan<- Slave) {
	bufc := bufio.NewReader(c)
	defer c.Close()

	//we first need to add current client to the channel
	//filling in the client structure
	NickName:=promptNick(c, bufc)
	if NickName =="client"{
		client := Client{
		conn:     c,
		nickname: NickName,
		ch:       make(chan string),
	}

	if strings.TrimSpace(client.nickname) == "" {
		io.WriteString(c, "Invalid Username\n")
		return
	}

	// Register user, our messageHandler is waiting on this channel
	//it populates the map
	addchanclient <- client

	//just a welcome message
	io.WriteString(c, fmt.Sprintf("Welcome, %s!\n\n", client.nickname))

	//We are now populating the other channel now
	//our message handler is waiting on this channel as well
	//it reads this message and copies to the individual channel of each Client in map
	// effectively the broadcast
	//msgchan <- fmt.Sprintf("New user %s has joined the chat room.\n", client.nickname)

	// another go routine whose purpose is to keep on waiting for user input
	//and write it with nick to the
	go client.ReadLinesInto(msgchan,removechanclient)

	//given a channel, writelines prints lines from it
	//we are giving here client.ch and this routine is for each client
	//so effectively each client is printitng its channel
	//to which our messagehandler has added messages for boroadcast
	client.WriteLinesFrom(client.ch ,removechanclient)
	} else if NickName == "slave"{

	slave := Slave{
	conn:     c,
	nickname: NickName,
	ch:       make(chan string),
	}

	if strings.TrimSpace(slave.nickname) == "" {
		io.WriteString(c, "Invalid Username\n")
		return
	}

	// Register user, our messageHandler is waiting on this channel
	//it populates the map
	addchanslave <- slave

	//just a welcome message
	io.WriteString(c, fmt.Sprintf("Welcome, %s!\n\n", slave.nickname))

	//We are now populating the other channel now
	//our message handler is waiting on this channel as well
	//it reads this message and copies to the individual channel of each Client in map
	// effectively the broadcast
	msgchanslave <- fmt.Sprintf("New user %s has joined the chat room.\n", slave.nickname)

	// another go routine whose purpose is to keep on waiting for user input
	//and write it with nick to the
	go slave.ReadLinesInto(msgchanslave)

	//given a channel, writelines prints lines from it
	//we are giving here client.ch and this routine is for each client
	//so effectively each client is printitng its channel
	//to which our messagehandler has added messages for boroadcast
	slave.WriteLinesFrom(slave.ch,removechanslave)

	}	else{
		io.WriteString(c, "Invalid Username\n")
		return
	}
	
}


func handleMessagesClient(msgchan <-chan string, msgchanslave <-chan string, addchanclient <-chan Client, removechanclient <-chan Client, addchanslave <-chan Slave, removechanslave <-chan Slave) {
	clients := make(map[net.Conn]chan<- string)
	slaves := make(map[net.Conn]chan<- string)
	for {
		select {
		case msg := <-msgchan:
			log.Printf("New message: %s", msg)
			if len(slaves)==0{
				fmt.Println("No slaves connected")
				continue;
			} else{
				fmt.Println(len(slaves))
			}
			for _, ch := range slaves {
				//go func(mch chan<- string) { mch <- "\033[1;33;40m" + msg + "\033[m" }(ch)
				//ch <- "\033[1;33;40m" + msg + "\033[m"
				ch <- msg
			}

		case msg := <-msgchanslave:
			log.Printf("New message: %s", msg)
			if len(clients)==0{
				fmt.Println("No clients connected")
				continue;
			}
			for _, ch := range clients {
				//go func(mch chan<- string) { mch <- "\033[1;33;40m" + msg + "\033[m" }(ch)
				//ch <- "\033[1;33;40m" + msg + "\033[m"
				ch <- msg
			}	
		case client := <-addchanclient:
			log.Printf("New client: %v\n", client.conn)
			clients[client.conn] = client.ch

		case client := <-removechanclient:
			log.Printf("Remove client: %v\n", client.conn)
			delete(clients,client.conn)

		case slave := <-addchanslave:
			log.Printf("New client: %v\n", slave.conn)
			slaves[slave.conn] = slave.ch

		case slave := <-removechanslave:
			log.Printf("Remove slave: %v\n", slave.conn)
			delete(slaves,slave.conn)


		}
	}
}

/*func handleMessagesSlave(msgchan <-chan string, addchanslave <-chan Slave, removechanslave <-chan Slave) {
	slaves := make(map[net.Conn]chan<- string)

	for {
		select {
		case msg := <-msgchan:
			log.Printf("New message: %s", msg)
			for _, ch := range slaves {
				//go func(mch chan<- string) { mch <- "\033[1;33;40m" + msg + "\033[m" }(ch)
				//ch <- "\033[1;33;40m" + msg + "\033[m"
				ch<-msg
			}
		case slave := <-addchanslave:
			log.Printf("New client: %v\n", slave.conn)
			slaves[slave.conn] = slave.ch

		case slave := <-removechanslave:
			log.Printf("Remove slave: %v\n", slave.conn)
			delete(slaves,slave.conn)


		}
	}
}*/
