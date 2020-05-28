package parallel

type empty struct{}
type WorkerPool chan empty

func NewWorkerPool(size int) WorkerPool {
	return make(chan empty, size)
}

func (p WorkerPool) Checkout() {
	p <- empty{}
}

func (p WorkerPool) Checkin() {
	<-p
}
