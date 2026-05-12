package strategies

import "designpatterns/lld/cointoss"

type Opposite struct{}

func (op *Opposite) Execute(systemOutcome cointoss.Outcome, userOutcome cointoss.Outcome) bool {
	return systemOutcome != userOutcome
}
