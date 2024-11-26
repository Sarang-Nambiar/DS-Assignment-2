package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"voting_protocol/node"
	"voting_protocol/utils"
)

// TODO: Make a terminal interface to start the token passing(ONLY FOR BOOTSTRAP)
// Add in the ability to input the number of nodes that would be requesting for the critical section

func main() {
	n := node.Node{
		Votes: 1, 
		VotesReceived: []node.Pointer{},
		Clock: 0,
		Queue: utils.NewPriorityQueue(),
		Lock: sync.Mutex{}, 
	}

	var nodesList map[int]string = utils.ReadNodesList()

	if len(nodesList) == 0 {
		n.ID = 0 // Set as bootstrap node
		n.IP = node.LOCALHOST + "8000"
		n.Network = make(map[int]string)
	} else {
		n.ID = len(nodesList)
		n.IP = node.LOCALHOST + strconv.Itoa(8000 + n.ID)	
		n.Network = make(map[int]string)
		for i := range nodesList {
			n.Network[i] = nodesList[i]
		}
	}

	go n.StartRPCServer()

	for i := range nodesList {
		message := node.Message{ID: n.ID, IP: n.IP}
		_, err := node.CallByRPC(nodesList[i], "Node.AddNode", message)
		if err != nil {
			fmt.Printf("[NODE-%d] Error occurred while adding node %d to the network: %s\n", n.ID, i, err)
		}
	}

	nodesList[n.ID] = n.IP

	jsonData, err := json.Marshal(nodesList)
	if err != nil {
		fmt.Println("Error occurred while marshalling nodesList: ", err)
	}

	err = ioutil.WriteFile("nodes-list.json", jsonData, os.ModePerm)

	if err != nil {
		fmt.Println("Error occurred while updating nodes-list.json: ", err)
	}

	var numRequests int
	if n.ID == 0 {
		fmt.Printf("[NODE-%d] Make sure all the nodes are up and running.\n", n.ID)
		fmt.Printf("[NODE-%d] How many nodes should request for CS: \n", n.ID)
		fmt.Scan(&numRequests)

		nodesList = utils.ReadNodesList()
		message := node.Message{NumRequests: numRequests}
		for i := 0; i < len(nodesList); i++ {
			go func(i int) {
				_, err := node.CallByRPC(nodesList[i], "Node.SetRequesting", message)
				if err != nil {
					fmt.Printf("[NODE-%d] Error occurred while setting the request flag for node %d: %s\n", n.ID, i, err)
				}
			}(i)
		}
	}
	
	// Start the token passing
	if n.ID == 0 {
		var answer string
		go func() {
			for {
				fmt.Printf("[NODE-%d] Do you want to start the request process? (y/n): ", n.ID)
				fmt.Scan(&answer)
				if answer == "y" {
					for i := range nodesList {
						go func(i int) {
							_, err := node.CallByRPC(nodesList[i], "Node.StartRequestProcess", node.Message{})
							if err != nil {
								fmt.Printf("[NODE-%d] Error occurred while starting the request process for node %d: %s\n", n.ID, i, err)
							}
						}(i)
					}
					break
				} else { 
					fmt.Printf("[NODE-%d] Waiting for all nodes to be ready...\n", n.ID)
				}
			}
		}()
	}

	go utils.CalculateTimeTaken(&n, numRequests)

	// Handling when the node fails or is shut down
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// For cleanup after the node is shut down
	go func() {
		<-sigChan
		fmt.Println("Shutting down...")

		// Remove the node from the list
		nodesList = utils.ReadNodesList()

		delete(nodesList, n.ID) // remove the element that left the network from the nodesList

		jsonData, err := json.Marshal(nodesList)
		err = ioutil.WriteFile("nodes-list.json", jsonData, os.ModePerm)
		if err != nil {
			fmt.Println("Error occurred while updating nodes-list.json: ", err)
		}
		os.Exit(0)
	}()

	select {}
}



