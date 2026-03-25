package awsprovider

import "fmt"

type Rds struct{}

func (rds *Rds) Query() { fmt.Println("Querying rds!") }
