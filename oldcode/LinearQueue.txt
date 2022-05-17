package req

type LinearQueue struct {
	list []*RequestSend
}

func (lq *LinearQueue) Add(rs *RequestSend) {
	lq.list = append(lq.list, rs)
}

func (lq *LinearQueue) Pop() *RequestSend {
	var x *RequestSend
	x, lq.list = lq.list[0], lq.list[1:]
	return x
}
