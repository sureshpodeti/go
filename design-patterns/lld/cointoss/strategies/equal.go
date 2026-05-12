package strategies

import (
	"designpatterns/lld/cointoss"
	"fmt"
)

type Equal struct{}

func (eq *Equal) Execute(systemOutcome cointoss.Outcome, userOutcome cointoss.Outcome) bool {
	fmt.Println(systemOutcome, userOutcome)
	return systemOutcome == userOutcome
}
