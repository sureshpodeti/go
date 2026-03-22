package statepattern

import "fmt"

type Draft struct {
	document *Document
}

// states: draft, submitted, approved, rejected
// actions: submit, approve, reject

func NewDraft(d *Document) *Draft {
	return &Draft{
		document: d,
	}
}

func (d *Draft) Submit() {
	fmt.Println("Document submitted!")
	d.document.ChangeState(d.document.submittedState)
}

func (d *Draft) Approve() {
	fmt.Println("Can't approve a Draft!")
}

func (d *Draft) Reject() {
	fmt.Println("Can't reject a Draft!")
}
