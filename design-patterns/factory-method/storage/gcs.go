package storage

import "fmt"

type GCS struct{}

func (gcs *GCS) Upload() {
	fmt.Println("Uploading file to GCS")
}
