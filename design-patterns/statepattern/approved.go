package statepattern

type Approved struct {
	document *Document
}

func NewApproved(doc *Document) *Approved {
	return &Approved{document: doc}
}

func (a *Approved) Submit() {

}

func (a *Approved) Approve() {

}

func (a *Approved) Reject() {

}
