package speed

type PoolManager struct {
	Pools []Pool
}

var (
	PM PoolManager
)

func (p *PoolManager) Start() {
	p.Pools = make([]Pool, 0, 5)
}
