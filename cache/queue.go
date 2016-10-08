package cache

var Q *Queue = &Queue{List: make([]string, 10)}

type Queue struct {
	List []string
}

func (q *Queue) Contains(e string) bool {
	for _, v := range q.List {
		if v == e {
			return true
		}
	}
	return false
}

func (q *Queue) Add(e string) *Queue {
	q.List = append(q.List, e)
	return q
}

func (q *Queue) Remove(e string) *Queue {
	indexS := []int{}
	for i, v := range q.List {
		if v == e {
			indexS = append(indexS, i)
		}
	}
	for _, v := range indexS {
		q.List = append(q.List[:v], q.List[v+1:]...)
	}
	return q
}
