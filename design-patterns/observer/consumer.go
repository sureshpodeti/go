package observer

import "fmt"

type Consumer struct {
	ID int
}

func NewConsumer(id int) *Consumer {
	return &Consumer{ID: id}
}

func (c *Consumer) Update() {
	fmt.Printf("Consumer %d received update\n", c.ID)
}
