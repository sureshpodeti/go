package factory

import "designpatterns/abstract-factory/cloudprovider"

type CloudFactory interface {
	CreateStorage() cloudprovider.Storage
	CreateDatabase() cloudprovider.Database
	CreateCompute() cloudprovider.Compute
}
