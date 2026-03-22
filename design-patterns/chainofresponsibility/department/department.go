package department

import "designpatterns/chainofresponsibility/model"

type Department interface {
	Executer(p *model.Patient)
}
