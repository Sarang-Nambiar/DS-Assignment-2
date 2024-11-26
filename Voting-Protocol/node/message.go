package node

type Message struct {
	Type string 
	ID int
	IP string
	ReqTime int
	Clock int
	NumRequests int
}