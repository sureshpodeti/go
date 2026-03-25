package abstractfactory

import "fmt"

type GCP struct{}

func (gcp *GCP) Upload() {
	fmt.Println("Uplo")
}
