package node

import (

)

type Message struct {
	Type string // Request, Reply, Release
	ID int
	IP string
	ReqTime int
	Clock int
	NumRequests int
}