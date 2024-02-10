package main

import (
	"Network-go/network/bcast"
	"Network-go/network/peers"
	//"fmt"
	//"os"
	//"time"
)

type msg struct {
	msg    string
	number uint
}

type con_status int
type tx_status int
type rx_status int

const (
	Disconnected con_status = iota
	Connecting
	Connected
)

//The communication FSM goes like this:
//Sender: *send* 0->Transmitted
//Reciever: *recieves and sends back* 0->Recieved
//Sender: *recieves echo* Transmitted -> Confirmed (if same)
//Reciever: *recieves confirmation* Recieved -> Confirmed_recieved

//Some "alive" statement every x ms also needs to be incorporated

const (
	Transmitted tx_status = iota
	Confirmed_transmitted
)

const (
	Recieved rx_status = iota
	Confirmed_recieved
)

type node struct {
	id         string //unique id for the network
	connection con_status
	transmit   tx_status
	recieve    rx_status
}

func transmit(tx_ch chan string) {
	//
}

func recieve(ch chan string, port int) {
	bcast.Receiver(port, ch)
}

func main() {

	//Check for active nodes on ip:port
	peerUpdateCh := make(chan peers.PeerUpdate)
	go peers.Receiver(30000, peerUpdateCh)

	//define the datatype we want to recive
	rx_ch := make(chan string)

	//recieve thread
	go recieve(rx_ch, 30000)

	//printing loop
	for {
		println(<-rx_ch)
	}
}
