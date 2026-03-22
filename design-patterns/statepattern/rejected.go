package statepattern

type Rejected struct {
	document *Document
}

func NewRejected(doc *Document) *Rejected {
	return &Rejected{document: doc}
}

func (r *Rejected) Submit() {

}

func (r *Rejected) Approve() {

}

func (r *Rejected) Reject() {

}
