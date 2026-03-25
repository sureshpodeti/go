package factory

import (
	"designpatterns/abstract-factory/cloudprovider"
	"designpatterns/abstract-factory/cloudprovider/awsprovider"
)

type AwsFactory struct{}

func (aws *AwsFactory) CreateStorage() cloudprovider.Storage {
	return &awsprovider.S3{}
}

func (aws *AwsFactory) CreateDatabase() cloudprovider.Database {
	return &awsprovider.Rds{}
}

func (aws *AwsFactory) CreateCompute() cloudprovider.Compute {
	return &awsprovider.Ec2{}
}
