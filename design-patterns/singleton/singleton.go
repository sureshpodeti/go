package singleton

import (
	"fmt"
	"sync"
)

var Instance *Singleton

type Singleton struct{}

var lock = &sync.Mutex{}

func GetInstance() *Singleton {
	if Instance == nil {
		lock.Lock()
		defer lock.Unlock()
		if Instance == nil {
			fmt.Println("Creating single instance now")
			Instance = &Singleton{}
		} else {
			fmt.Println("Single instance is already created")
		}
	} else {
		fmt.Println("Single instance is already created")
	}
	return Instance
}
