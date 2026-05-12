package cointoss

import (
	"fmt"
	"log"
)

type InputTaker struct{}

func (inpt *InputTaker) TakeInput() (Outcome, error) {
	var char string
	log.Print("Input Your Guess (H/ T) - ")

	fmt.Scanf("%s", &char)

	var outcome Outcome
	var err error = nil
	switch char {
	case "H", "h":
		outcome = HEAD
	case "T", "t":
		outcome = TAIL
	default:
		err = fmt.Errorf("Invalid input")
	}

	return outcome, err
}
