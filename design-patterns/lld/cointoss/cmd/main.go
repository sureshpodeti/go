package main

import (
	"designpatterns/lld/cointoss"
	"designpatterns/lld/cointoss/strategies"
)

func main() {
	coinTosser := &cointoss.CoinTosser{}
	inputTaker := &cointoss.InputTaker{}

	strategy := &strategies.Equal{}
	evaluator := &cointoss.Evaluator{}
	evaluator.SetStrategy(strategy)

	game := cointoss.NewGame(coinTosser, inputTaker, evaluator)

	game.Play()
}
