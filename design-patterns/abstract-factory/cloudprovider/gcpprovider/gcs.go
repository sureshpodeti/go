package gcpprovider

import "fmt"

type Gcs struct{}

func (gcs *Gcs) Upload() { fmt.Println("Uploading to Gcs") }
