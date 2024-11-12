package node

import (

)

type Message struct {
	ID int // ID of the node
	IP string // Source IP
	ReqTime int // timestamp assigned to the token
	Clock int
}