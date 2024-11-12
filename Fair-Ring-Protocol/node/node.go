package node

import (
	"fmt"
	"math/rand"
	"net"
	"net/rpc"
	"os"
	"sync"
	"time"
)

 type Node struct {
	ID int
	IP string
	Successor string // IP of the successor of the node
	Clock int
	Request bool // boolean to check if the node will concurrently request for the token
	Flag bool // boolean to keep calculating whether to request for the token or not
	ReqTime int // timestamp at which the node requests for the token
	Lock sync.Mutex
 }

 const (
	LOCALHOST = "127.0.0.1:"
 )

 // Dummy critical section function
 func (n *Node) CriticalSection() {
	n.Lock.Lock()
	defer n.Lock.Unlock()
	// Simulate entering the critical section
	fmt.Printf("[NODE-%d] Entering the critical section\n", n.ID)
	time.Sleep(2 * time.Second)
	fmt.Printf("[NODE-%d] Completed the critical section\n", n.ID)
 }

 // Function to start the RPC server
 func (n *Node) StartRPCServer() {
	rpc.Register(n)

	listener, err := net.Listen("tcp", n.IP)
	if err != nil {
		fmt.Printf("[NODE-%d] could not start listening: %s\n", n.ID, err)
		os.Exit(1)
	}
	defer listener.Close()

	fmt.Printf("[NODE-%d] Node is running on %s\n", n.ID, n.IP)

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("[NODE-%d] accept error: %s\n", n.ID, err)
			continue
		}
		go rpc.ServeConn(conn)
	}
 }

 // Initialize the token passing
 func (n *Node) StartTokenPassing() {
	n.Clock += 1
	message := Message{ID: n.ID, Clock: n.Clock, ReqTime: -1}

	// Send the token to the successor concurrently
	go func (){
		_, err := CallByRPC(n.Successor, "Node.ReceiveToken", message)

		if err != nil {
			fmt.Printf("[NODE-%d] Error occurred while sending token: %s\n", n.ID, err)
			return 
		}
	}()
 }

 // Function to receive the token
 func (n *Node) ReceiveToken(message Message, reply *Message) error {
	time.Sleep(1 * time.Second)
	fmt.Printf("[NODE-%d] Received token from NODE-%d\n", n.ID, message.ID)
	n.Clock = max(n.Clock, message.Clock) + 1

	if !n.Request {
		n.Request = n.isRequesting()
	} 

	if n.Request {

		// Update the logical clock
		if n.ReqTime == -1 {
			n.ReqTime = n.Clock
		}

		fmt.Printf("[NODE-%d] Requesting for the token at timestamp-%d\n", n.ID, n.ReqTime)

		// check the values of the timestamp from the message
		if message.ReqTime == -1 {
			message.ReqTime = n.ReqTime

		}else if message.ReqTime == n.ReqTime {
			// Run the critical section
			n.CriticalSection()
			n.Request = false // Reset the request Flag
			n.Flag = false // Reset the Flag
			message.ReqTime = -1 // Reset the timestamp
			n.ReqTime = -1 // Reset the timestamp

		} else if n.ReqTime < message.ReqTime {
			message.ReqTime = n.ReqTime
		}
	}

	message.ID = n.ID
	message.Clock = n.Clock + 1

	// Send the token to the successor concurrently
	go func() {
		_, err := CallByRPC(n.Successor, "Node.ReceiveToken", message)
		if err != nil {
			fmt.Printf("[NODE-%d] Error occurred while sending token: %s\n", n.ID, err)
			return
		}
	}()

	return nil 
 }

 // Function to set the successor of the node
 func (n *Node) SetSuccessor(message Message, reply *Message) error {
	n.Successor = message.IP
	fmt.Printf("[NODE-%d] Successor set to %s\n", n.ID, n.Successor)
	return nil
}

// Function to check if the node is going to request the token or not
func (n *Node) isRequesting() bool {
	if n.Flag {
		rand.Seed(time.Now().UnixNano()) // Making sure this is random using a unique seed
		return rand.Intn(2) == 1 // Generates random number from 0 to 1
	}
	return false
}

// Utility function to call RPC methods
func CallByRPC(IP string, method string, message Message) (Message, error) {
	client, err := rpc.Dial("tcp", IP)
	if err != nil {
		return Message{}, fmt.Errorf("error in dialing: %s", err)
	}
	defer client.Close()

	var reply Message
	err = client.Call(method, message, &reply)
	if err != nil {
		return Message{}, fmt.Errorf("error in calling %s: %s", method, err)
	}
	return reply, nil
}