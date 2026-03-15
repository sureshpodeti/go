package contexts

import "errors"

type Strategy interface {
	Do(num1, num2 int) int
}

type ArithmeticContext struct {
	strategy Strategy
}

func NewArithmeticContext(s Strategy) *ArithmeticContext {
	return &ArithmeticContext{strategy: s}
}

func (ar *ArithmeticContext) SetStrategy(strategy Strategy) {
	ar.strategy = strategy
}

func (ar *ArithmeticContext) Execute(num1, num2 int) (int, error) {
	if ar.strategy == nil {
		return 0, errors.New("strategy not set")
	}
	return ar.strategy.Do(num1, num2), nil
}
