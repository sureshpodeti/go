package main

import (
	"context"
	"fmt"
	"time"
)

func sleep(ctx context.Context) {

	select {
	case <-time.After(time.Second * 5):
		fmt.Println("hello")
	case <-ctx.Done():
		fmt.Println("context cancelled", ctx.Err().Error())
	}
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*4)
	defer cancel()

	sleep(ctx)

}
