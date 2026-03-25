package storage

import "fmt"

type Local struct{}

func (local *Local) Upload() {
	fmt.Println("Uploading file to local storage")
}
