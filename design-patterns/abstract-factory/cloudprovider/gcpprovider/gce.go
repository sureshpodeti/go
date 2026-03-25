package gcpprovider

import "fmt"

type Gce struct{}

func (gce *Gce) Run() { fmt.Println("Running Gce instance") }
