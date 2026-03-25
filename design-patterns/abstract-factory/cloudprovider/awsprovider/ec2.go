package awsprovider

import "fmt"

type Ec2 struct{}

func (ec2 *Ec2) Run() { fmt.Println("Running Ec2 instance") }
