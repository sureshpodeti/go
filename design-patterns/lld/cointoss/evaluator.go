package cointoss

type Strategy interface {
	Execute(systemOutcome Outcome, userOutcome Outcome) bool
}
type Evaluator struct {
	strategy Strategy
}

func (eval *Evaluator) SetStrategy(s Strategy) {
	eval.strategy = s
}

func (eval *Evaluator) Evaluate(systemOutcome Outcome, userOutcome Outcome) bool {
	return eval.strategy.Execute(systemOutcome, userOutcome)
}
