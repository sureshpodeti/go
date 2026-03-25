package storage

import "fmt"

type S3 struct{}

func (s3 *S3) Upload() {
	fmt.Println("Uploading file  to S3")
}
