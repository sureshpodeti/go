package cointoss

import (
	"log"
	"math/rand"
	"time"
)

type CoinTosser struct{}

func (ct *CoinTosser) TossCoin() Outcome {
	rand.NewSource(time.Now().UnixNano())
	n := rand.Intn(2)
	var outcome Outcome
	switch n {
	case 0:
		outcome = HEAD
	case 1:
		outcome = TAIL
	}
	log.Println("Coin tossed!")
	return outcome
}
