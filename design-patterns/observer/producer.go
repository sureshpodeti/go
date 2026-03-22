package observer

type Observer interface {
	Update()
}

type Producer struct {
	observers []Observer
}

func NewProducer() *Producer {
	return &Producer{}
}

func (p *Producer) Register(o Observer) {
	p.observers = append(p.observers, o)
}

func (p *Producer) Deregister(o Observer) {
	for i, obs := range p.observers {
		if obs == o {
			p.observers = append(p.observers[:i], p.observers[i+1:]...)
			break
		}
	}
}

func (p *Producer) Notify() {
	for _, o := range p.observers {
		o.Update()
	}
}
