package main

import "designpatterns/observer"

func main() {
	producer := observer.NewProducer()
	consumer1 := observer.NewConsumer(1)
	consumer2 := observer.NewConsumer(2)
	consumer3 := observer.NewConsumer(3)

	producer.Register(consumer1)
	producer.Register(consumer2)
	producer.Register(consumer3)

	producer.Notify()
}
