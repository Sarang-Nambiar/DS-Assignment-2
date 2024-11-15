package node

type Item struct {
	ID        int
	TimeStamp int
}

type PriorityQueue []Item

// Size of the priority queue
func (pq PriorityQueue) Len() int { return len(pq) }

func (pq PriorityQueue) Less(i, j int) bool {
	if pq[i].TimeStamp == pq[j].TimeStamp {
		return pq[i].ID < pq[j].ID
	}
	return pq[i].TimeStamp < pq[j].TimeStamp
}

func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
}

func (pq *PriorityQueue) Push(x interface{}) {
	item := x.(Item)
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