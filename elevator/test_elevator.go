package elevator

import (
	"fmt"
	"time"
)

func F_Test_elevator() {
	c_receiveRequest := make(chan T_Request)
	c_sendRequest := make(chan T_Request)
	go F_RunElevator(c_receiveRequest, c_sendRequest)

	for {
		select {
		case a := <-c_sendRequest:
			if a.Calltype == CAB {
				fmt.Printf("Dummynode: Mottok cabrequest fra floor %v+\n", a.Floor)
			} else {
				fmt.Printf("Dummynode: Mottok hallrequest fra floor %v+\n", a.Floor)
			}
			time.Sleep(3 * time.Second)
			fmt.Printf("Dummynode: Sender request til heis\n")
			c_receiveRequest <- a
		}
	}
}
