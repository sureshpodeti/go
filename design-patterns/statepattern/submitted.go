package statepattern

type Submitted struct {
	document *Document
}

func NewSubmitted(doc *Document) *Submitted {
	return &Submitted{
		document: doc,
	}
}

func (s *Submitted) Submit() {

}

func (s *Submitted) Approve() {

}

func (s *Submitted) Reject() {

}
