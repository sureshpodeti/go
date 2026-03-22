package statepattern

type State interface {
	Submit()
	Approve()
	Reject()
}
type Document struct {
	state State

	draftState, submittedState, approvedState, rejectedState State
}

func NewDocument() *Document {

	document := &Document{}
	draftState := NewDraft(document)
	submittedState := NewSubmitted(document)
	approvedState := NewApproved(document)
	rejectedState := NewRejected(document)

	document.draftState = draftState
	document.submittedState = submittedState
	document.approvedState = approvedState
	document.rejectedState = rejectedState

	document.state = draftState
	return document
}

func (d *Document) ChangeState(s State) {
	d.state = s
}

func (d *Document) Submit() {
	d.state.Submit()
}

func (d *Document) Approve() {
	d.state.Approve()
}

func (d *Document) Reject() {
	d.state.Reject()
}
