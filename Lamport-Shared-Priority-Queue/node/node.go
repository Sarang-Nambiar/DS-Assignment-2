package node

import (
	"container/heap"
	"fmt"
	"net"
	"net/rpc"
	"os"
	"sync"
	"time"
)

type Node struct {
	ID int
	IP string
	Queue *PriorityQueue // Containing the list of IP addresses to contact 
	NumVotes int // Number of votes the node has
	Clock int // Lamport clock
	Request bool // If the node is requesting the token
	ReqTime int // Request timestamp
	Network map[int]string // Map of the network
	Finished []bool // If the node has finished
	Lock sync.Mutex
}

const (
	LOCALHOST = "127.0.0.1:"
	ACK = "ACK"
	REPLY = "REPLY"
	REQUEST = "REQUEST"	
	RELEASE = "RELEASE"
)

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

func (n *Node) StartRequestProcess(message Message, reply *Message) error {
	if n.Request {
		n.Clock++
		n.ReqTime = n.Clock

		// Add the request to the queue
		n.Lock.Lock()
		heap.Push(n.Queue, Item{ID: n.ID, TimeStamp: n.ReqTime})
		fmt.Printf("[NODE-%d] Added node %d with request timestamp %d to the queue. New Queue: %v\n", n.ID, n.ID, n.ReqTime, n.Queue)
		n.Lock.Unlock()

		for i := range n.Network {
			_, err := CallByRPC(n.Network[i], "Node.ReceiveMessage", Message{Type: REQUEST, ID: n.ID, ReqTime: n.ReqTime, Clock: n.Clock})
			if err != nil {
				return fmt.Errorf("[NODE-%d] Error occurred while sending a request to node %d: %s\n", n.ID, i, err)
			}
			n.Clock++
		}
	}
	return nil
}

// Dummy critical section function
func (n *Node) CriticalSection() {
	n.Lock.Lock()
	defer n.Lock.Unlock()
	// Simulate entering the critical section
	fmt.Printf("[NODE-%d] Entering the critical section\n", n.ID)
	time.Sleep(2 * time.Second)
	fmt.Printf("[NODE-%d] Completed the critical section\n", n.ID)

	// Notify Bootstrap node when the critical section is completed
	_, err := CallByRPC(LOCALHOST + "8000", "Node.NotifyFinished", Message{ID: n.ID})
	if err != nil {
		fmt.Printf("[NODE-%d] Error occurred while notifying the bootstrap node: %s\n", n.ID, err)
	}
}

func (n *Node)NotifyFinished(message Message, reply *Message) error {
	n.Finished[message.ID] = true
	return nil
}

// Handle the different types of messages
func (n *Node)ReceiveMessage(message Message, reply *Message) error {
	n.Lock.Lock()
	n.Clock = max(n.Clock, message.Clock) + 1
	n.Lock.Unlock()
	
	time.Sleep(1 * time.Second)
	switch message.Type {
	case REQUEST:
		top := n.Queue.Peek()
		if top == nil {
			// Direct send a reply since there is no other request in the queue
			n.Clock++
			fmt.Printf("[NODE-%d] Received a request from node %d. Sending a reply directly since the queue is empty\n", n.ID, message.ID)
			_, err := CallByRPC(n.Network[message.ID], "Node.ReceiveMessage", Message{Type: REPLY, ID: n.ID, Clock: n.Clock})
			if err != nil {
				fmt.Printf("[NODE-%d] Error occurred while sending a reply to node %d: %s\n", n.ID, message.ID, err)
			}
		} else {
			if top.(Item).TimeStamp > message.ReqTime {
				// Directly send a reply since the request is earlier than the top of the queue
				n.Clock++
				fmt.Printf("[NODE-%d] Received a request from node %d. Sending a reply directly since the request is earlier than the top of the queue\n", n.ID, message.ID)
				_, err := CallByRPC(n.Network[message.ID], "Node.ReceiveMessage", Message{Type: REPLY, ID: n.ID, Clock: n.Clock})
				if err != nil {
					fmt.Printf("[NODE-%d] Error occurred while sending a reply to node %d: %s\n", n.ID, message.ID, err)
				}
			} else if top.(Item).TimeStamp == message.ReqTime { // Case when the logical clocks could be the same
				// Compare the IDs of the nodes
				fmt.Printf("[NODE-%d] Received a request from node %d with same request timestamps. Comparing the IDs of the nodes\n", n.ID, message.ID)
				if top.(Item).ID > message.ID {
					// Directly send a reply since the message ID is smaller than the queue ID
					fmt.Printf("[NODE-%d] Sending a reply to node %d since the ID is smaller than the queue ID\n", n.ID, message.ID)
					_, err := CallByRPC(n.Network[message.ID], "Node.ReceiveMessage", Message{Type: REPLY, ID: n.ID, Clock: n.Clock})
					if err != nil {
						fmt.Printf("[NODE-%d] Error occurred while sending a reply to node %d: %s\n", n.ID, message.ID, err)
					}
				} else {
					// Add the request to the queue
					n.Lock.Lock()
					heap.Push(n.Queue, Item{ID: message.ID, TimeStamp: message.ReqTime})
					fmt.Printf("[NODE-%d] Added node %d with timestamp %d to the queue. New Queue: %v\n", n.ID, message.ID, message.ReqTime, n.Queue)
					n.Lock.Unlock()
				}
			} else {
				// Add the request to the queue
				n.Lock.Lock()
				heap.Push(n.Queue, Item{ID: message.ID, TimeStamp: message.ReqTime})
				fmt.Printf("[NODE-%d] Added node %d with timestamp %d to the queue. New Queue: %v\n", n.ID, message.ID, message.ReqTime, n.Queue)
				n.Lock.Unlock()
			}
		}
	
	case REPLY:
		fmt.Printf("[NODE-%d] Received a reply from node %d\n", n.ID, message.ID)
		n.NumVotes++ 
		n.RequestCriticalSection()
	}
	return nil
}

// Function to add a new node to the network
func (n *Node)AddNode(message Message, reply *Message) error {
	n.Network[message.ID] = message.IP
	fmt.Printf("[NODE-%d] Added node %d to the network. New network: %v\n", n.ID, message.ID, n.Network)
	*reply = Message{Type: ACK}
	return nil
}

func (n *Node)RequestCriticalSection() {
	if n.NumVotes == len(n.Network) {
		fmt.Printf("[NODE-%d] Received all the votes: %d\n", n.ID, n.NumVotes)
		n.CriticalSection()

		// Reset the node's request status
		n.NumVotes = 0
		n.Request = false

		n.Lock.Lock()
		defer n.Lock.Unlock()

		heap.Pop(n.Queue)
		queueLen := n.Queue.Len() // fixes the length of the queue since the length of the queue updates after every pop
		for i := 0; i < queueLen; i++ {
			item := heap.Pop(n.Queue)

			n.Clock++
			msg := Message{Type: REPLY, ID: n.ID, Clock: n.Clock}
			_, err := CallByRPC(n.Network[item.(Item).ID], "Node.ReceiveMessage", msg)
			if err != nil {
				fmt.Printf("[NODE-%d] Error occurred while sending a reply message to node %d: %s\n", n.ID, item.(Item).ID, err)
			}
			fmt.Printf("[NODE-%d] Sent a reply message to node %d\n", n.ID, item.(Item).ID)
		}
	}
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
