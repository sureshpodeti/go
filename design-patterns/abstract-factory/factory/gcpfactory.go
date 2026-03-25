package factory

import (
	"designpatterns/abstract-factory/cloudprovider"
	"designpatterns/abstract-factory/cloudprovider/gcpprovider"
)

type GcpFactory struct{}

func (gcp *GcpFactory) CreateStorage() cloudprovider.Storage {
	return &gcpprovider.Gcs{}
}

func (gcp *GcpFactory) CreateDatabase() cloudprovider.Database {
	return &gcpprovider.CloudSQL{}
}

func (gcp *GcpFactory) CreateCompute() cloudprovider.Compute {
	return &gcpprovider.Gce{}
}
