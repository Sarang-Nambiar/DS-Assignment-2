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
	ID    int
	IP    string
	VotesReceived []Pointer // List of nodes that have voted for the node
	Votes int // Number of votes the node has to send to other nodes
	Finished []bool
	PrevReq Pointer
	Queue *PriorityQueue // contains all the nodes that have requested for the critical section after the vote from the node was sent to another node.
	Network map[int]string // Contains the list of nodes in the network
	Clock int
	Request bool // whether the node should request for the critical section
	isFinished bool
	ReqTime int
	Lock sync.Mutex
}

const (
	LOCALHOST = "127.0.0.1:"
	ACK = "ACK"
	DENY = "DENY"
	VOTE = "VOTE"
	REQUEST = "REQUEST"	
	RESCIND_VOTE = "RESCIND_VOTE"
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

		// Send a CS request to all the nodes in the network
		for i := range n.Network {
			// concurrently start requesting the critical section
			go func() {
				fmt.Printf("[NODE-%d] Sending a request to node %d\n", n.ID, i)
				_, err := CallByRPC(n.Network[i], "Node.ReceiveMessage", Message{Type: REQUEST, ID: n.ID, IP: n.IP, ReqTime: n.ReqTime, Clock: n.Clock})
				if err != nil {
					fmt.Printf("[NODE-%d] Error occurred while sending a request to node %d: %s\n", n.ID, i, err)
				}
			}()
			n.Clock++
		}
	}
	return nil
}

// Dummy critical section function
func (n *Node) CriticalSection() {
	// Simulate entering the critical section
	fmt.Printf("[NODE-%d] Entering the critical section\n", n.ID)
	time.Sleep(2 * time.Second)
	fmt.Printf("[NODE-%d] Completed the critical section\n", n.ID)

	// Notify Bootstrap node when the critical section is completed
	n.Clock++
	n.Request = false // Reset the flag
	n.isFinished = true

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
	n.Clock = max(n.Clock, message.Clock) + 1

	time.Sleep(1 * time.Second)

	switch message.Type {
	case REQUEST:
		// Two cases:
		// 1. If the node has not voted for the CS of any other nodes, then vote for the requesting node
		// 2. If the node has voted for the CS of another node, then add the requesting node to the queue
		//   2.1 If the requesting node has a higher request timestamp, then add the requesting node to the queue
		//   2.2 If the requesting node has a lower request timestamp, then send a RESCIND_VOTE message to the previous node and add the previous node to the queue.
		fmt.Printf("[NODE-%d] Received a request from node %d\n", n.ID, message.ID)
		if n.Votes > 0 {

			n.Votes-- // Voting for the requesting node
			n.PrevReq = Pointer{ID: message.ID, IP: message.IP, ReqTime: message.ReqTime}

			n.Clock++
			fmt.Printf("[NODE-%d] Sending a vote to node %d\n", n.ID, message.ID)
			_, err := CallByRPC(message.IP, "Node.ReceiveMessage", Message{Type: VOTE, ID: n.ID, IP: n.IP, Clock: n.Clock})
			if err != nil {
				fmt.Printf("[NODE-%d] Error occurred while sending a vote to node %d: %s\n", n.ID, message.ID, err)
			}
			return nil
		} else {
			heap.Push(n.Queue, Pointer{ID: message.ID, IP: message.IP, ReqTime: message.ReqTime})
			fmt.Printf("[NODE-%d] Added node %d to the queue. New Queue: %v\n", n.ID, message.ID, n.Queue)
			
			if n.PrevReq.ReqTime > message.ReqTime {
				n.RescindVote(message)

			} else if n.PrevReq.ReqTime == message.ReqTime {
				if n.PrevReq.ID > message.ID {
					n.RescindVote(message)
				}
			} 
		}
	
		// For some reason the releasing of all the votes is not taking place, moreover, the addition of another vote messes it up
		// Probably some lock issue
		// Maybe make the release a go routine
	case VOTE:

		n.VotesReceived = append(n.VotesReceived, Pointer{ID: message.ID, IP: message.IP})
		fmt.Printf("[NODE-%d] Received a vote from node %d. Votes received: %v\n", n.ID, message.ID, n.VotesReceived)

		// Check if the node has received a majority of the votes
		n.Lock.Lock()
		if !n.isFinished && (len(n.VotesReceived) >= ((len(n.Network) + 1) / 2) + 1) { // Majority of the votes received
			fmt.Printf("[NODE-%d] Majority of the votes received. Entering the critical section\n", n.ID)
			n.CriticalSection() // Execute the CS

			// concurrently release all the votes
			go n.sendRelease()
		} else {
			if n.isFinished { // Send release to all the incoming votes which are not yet released after execution of CS
				go n.sendRelease()
			}
		}
		n.Lock.Unlock()	

	case RELEASE:

		fmt.Printf("[NODE-%d] Received a release from node %d\n", n.ID, message.ID)
		n.Votes = 1
		n.PrevReq = Pointer{}

		if n.Queue.Len() > 0 {
			n.Votes-- // Voting for the requesting node
			n.PrevReq = heap.Pop(n.Queue).(Pointer)
			n.Clock++

			fmt.Printf("[NODE-%d] Sending a vote to node %d\n", n.ID, n.PrevReq.ID)
			_, err := CallByRPC(n.PrevReq.IP, "Node.ReceiveMessage", Message{Type: VOTE, ID: n.ID, IP: n.IP, Clock: n.Clock})
			if err != nil {
				fmt.Printf("[NODE-%d] Error occurred while sending a vote to node %d: %s\n", n.ID, n.PrevReq.ID, err)
			}
		}
	
	case RESCIND_VOTE:
		fmt.Printf("[NODE-%d] Received a rescind vote from node %d\n", n.ID, message.ID)
	
		if n.isFinished {
			fmt.Printf("[NODE-%d] The node %d has already entered the critical section. Sending a DENY message for the rescind request\n", n.ID, message.ID)
			*reply = Message{Type: DENY}
			return nil
		}

		element := Pointer{ID: message.ID, IP: message.IP}
		if !Contains(n.VotesReceived, element) {
			fmt.Printf("[NODE-%d] Current node does not contain the node %d in the votes received list\n", n.ID, message.ID)
			*reply = Message{Type: DENY}
			return nil
		}
		
		// Remove the node from the votes received slice
		n.VotesReceived = Remove(n.VotesReceived, element)	
		fmt.Printf("[NODE-%d] Removed node %d from the votes received list. New list: %v\n", n.ID, message.ID, n.VotesReceived)	
	}
	*reply = Message{Type: ACK} 
	return nil
}

// Function to add a new node to the network
func (n *Node)AddNode(message Message, reply *Message) error {
	n.Network[message.ID] = message.IP
	*reply = Message{Type: ACK}
	return nil
}

func (n *Node) sendRelease() {
	votesList := n.VotesReceived // create a copy so that any changes in length do not affect the loop
	n.VotesReceived = []Pointer{} // Reset the votes received list
	for i := range votesList {
		n.Clock++
		fmt.Printf("[NODE-%d] Sending a release to node %d\n", n.ID, votesList[i].ID)
		_, err := CallByRPC(votesList[i].IP, "Node.ReceiveMessage", Message{Type: RELEASE, ID: n.ID, Clock: n.Clock})
		if err != nil {
			fmt.Printf("[NODE-%d] Error occurred while sending a release to node %d: %s\n", n.ID, votesList[i].ID, err)
		}
	}
}

// Function to rescind the vote
func (n *Node)RescindVote(message Message) {

	fmt.Printf("[NODE-%d] Sending a rescind vote to node %d to vote for node %d instead.\n", n.ID, n.PrevReq.ID, message.ID)

	n.Clock++
	reply, err := CallByRPC(n.PrevReq.IP, "Node.ReceiveMessage", Message{Type: RESCIND_VOTE, ID: n.ID, IP: n.IP, Clock: n.Clock})
	if err != nil {
		fmt.Printf("[NODE-%d] Error occurred while sending a rescind vote to node %d: %s\n", n.ID, message.ID, err)
	}

	if reply.Type == ACK {// If the previous node has accepted the RESCIND_VOTE message
		
		// Store a copy of the previous request
		prevRequest := n.PrevReq
		n.Clock++
		_, err := CallByRPC(n.IP, "Node.ReceiveMessage", Message{Type: RELEASE, ID: n.PrevReq.ID, IP: n.PrevReq.IP, Clock: n.Clock})
		if err != nil {
			fmt.Printf("[NODE-%d] Error occurred while sending a release to node %d: %s\n", n.ID, message.ID, err)
		}

		// Add the PrevReq to the queue after the release message has been sent
		heap.Push(n.Queue, Pointer{ID: prevRequest.ID, IP: prevRequest.IP, ReqTime: prevRequest.ReqTime})
		fmt.Printf("[NODE-%d] Added node %d to the queue. New Queue: %v\n", n.ID, prevRequest.ID, n.Queue)
	}
}

// Function to decide whether the node requests for vote or not
func (n *Node) SetRequesting(message Message, reply *Message) error {
	n.Request = n.ID < message.NumRequests
	if n.Request {
		fmt.Printf("[NODE-%d] Node will request for the critical section\n", n.ID)
	} else {
		fmt.Printf("[NODE-%d] Node will not request for the critical section\n", n.ID)
	}
	*reply = Message{Type: ACK}
	return nil
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

// Function to remove an element from a slice
func Remove(slice []Pointer, element Pointer) []Pointer {
	var j int
	for i := 0; i < len(slice); i++ {
		if slice[i].IP == element.IP && slice[i].ID == element.ID {
			j = i
			break
		}
	}
    return append(slice[:j], slice[j+1:]...)
}

func Contains(slice []Pointer, element Pointer) bool {
	for _, v := range slice {
		if v.IP == element.IP && v.ID == element.ID {
			return true
		}
	}
	return false
}