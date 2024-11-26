package node

type Pointer struct {
	ID int
	IP string
	ReqTime int
}

type PriorityQueue []Pointer

// Size of the priority queue
func (pq PriorityQueue) Len() int { return len(pq) }

func (pq PriorityQueue) Less(i, j int) bool {
	if pq[i].ReqTime == pq[j].ReqTime {
		return pq[i].ID < pq[j].ID
	}
	return pq[i].ReqTime < pq[j].ReqTime
}

func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
}

func (pq *PriorityQueue) Push(x interface{}) {
	item := x.(Pointer)
	*pq = append(*pq, item)
}

func (pq *PriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	*pq = old[0 : n-1]
	return item
}

func (pq PriorityQueue) Peek() interface{} {
	if len(pq) == 0 {
		return nil
	}
	return pq[0]
}