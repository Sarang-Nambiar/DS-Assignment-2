package main

import (
	"encoding/json"
	"fair_ring/node"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"
)

func main() {
	n := node.Node{
		Lock: sync.Mutex{}, 
		Request: false, 
		Clock: 0, 
		ReqTime: -1,
	}

	var nodesList map[int]string = readNodesList()

	if len(nodesList) == 0 {
		n.ID = 0 // Set as bootstrap node
		n.IP = node.LOCALHOST + "8000"
		n.Successor = n.IP
	} else {
		n.ID = len(nodesList)
		n.IP = node.LOCALHOST + strconv.Itoa(8000 + n.ID)	
		n.Successor = nodesList[0] // Set the successor of the last node to the first node

		message := node.Message{ID: n.ID, IP: n.IP}
		
		_, err := node.CallByRPC(nodesList[n.ID - 1], "Node.SetSuccessor", message)
		if err != nil {
			fmt.Println("Error occurred while setting the successor: ", err)
		}
	}

	go n.StartRPCServer()

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

		nodesList = readNodesList()
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
			fmt.Printf("[NODE-%d] Make sure that all the required nodes are up.\n", n.ID)
			for {
				fmt.Printf("[NODE-%d] Do you want to start the token passing? (y/n): ", n.ID)
				fmt.Scan(&answer)
				if answer == "y" {
					n.StartTokenPassing()
					break
				} else { 
					fmt.Printf("[NODE-%d] Waiting for all nodes to be ready...\n", n.ID)
				}
			}
		}()
	}

	// Caclculate the time taken
	go calculateTimeTaken(&n, numRequests)

	// Handling when the node fails or is shut down
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// For cleanup after the node is shut down
	go func() {
		<-sigChan
		fmt.Println("Shutting down...")

		// Remove the node from the list
		nodesList = readNodesList()

		delete(nodesList, n.ID) // remove the element that left the network from the nodesList

		jsonData, err := json.Marshal(nodesList)
		err = ioutil.WriteFile("nodes-list.json", jsonData, os.ModePerm)
		if err != nil {
			fmt.Println("Error occurred while updating nodes-list.json: ", err)
		}
		os.Exit(0)
	}()
	select{}
}

func readNodesList() map[int]string {
	jsonFile, err := os.Open("nodes-list.json")
	if err != nil {
		fmt.Println("Error opening nodes-list.json file:", err)
	}
	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)

	var nodesList map[int]string

	json.Unmarshal(byteValue, &nodesList) // Puts the byte value into the nodesList map

	return nodesList
}

// Calculate the time taken from the first node to request to the last node to exist the critical section
func calculateTimeTaken(n *node.Node, numRequests int) {
	startTime := time.Now()

	if n.ID == 0 {
		n.Finished = make([]bool, numRequests)
		for {
			if all(n.Finished) {
				fmt.Printf("Time taken for all nodes to exit the critical section: %v\n", time.Since(startTime))
				break
			}
		}
	}
}

func all(arr []bool) bool {
	for _, v := range arr {
		if !v {
			return false
		}
	}
	return true
}
