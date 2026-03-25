package gcpprovider

import "fmt"

type CloudSQL struct{}

func (csql *CloudSQL) Query() { fmt.Println("Querying CloudSQL") }
