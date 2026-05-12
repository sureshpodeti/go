package cointoss

import (
	"fmt"
	"log"
)

type Game struct {
	coinTosser *CoinTosser
	inputTaker *InputTaker
	evaluator  *Evaluator
}

func NewGame(ct *CoinTosser, inp *InputTaker, eval *Evaluator) *Game {
	return &Game{
		coinTosser: ct,
		inputTaker: inp,
		evaluator:  eval,
	}

}

func (gameOrch *Game) Play() {
	// toss coin
	systemOutcome := gameOrch.coinTosser.TossCoin()
	// take input
	playerInput, err := gameOrch.inputTaker.TakeInput()

	for err != nil {
		fmt.Println("Error - ", err)
		playerInput, err = gameOrch.inputTaker.TakeInput()
	}
	// evaluate
	outcome := gameOrch.evaluator.Evaluate(systemOutcome, playerInput)
	//print result
	var result Result
	switch outcome {
	case true:
		result = WIN
	case false:
		result = LOST
	}

	log.Printf("Outcome  - %s ", result)

}
