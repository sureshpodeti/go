package main

import "designpatterns/abstract-factory/factory"

func main() {
	cloudProvider := "aws"

	var cloudFactory factory.CloudFactory

	switch cloudProvider {
	case "aws":
		cloudFactory = &factory.AwsFactory{}
	case "gcp":
		cloudFactory = &factory.GcpFactory{}
	}

	storage := cloudFactory.CreateStorage()
	database := cloudFactory.CreateDatabase()
	compute := cloudFactory.CreateCompute()

	storage.Upload()
	database.Query()
	compute.Run()
}
