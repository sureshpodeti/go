package department

import (
	"designpatterns/chainofresponsibility/model"
	"fmt"
)

type Reception struct {
	Next Department
}

func (r *Reception) Executer(p *model.Patient) {
	if p.RegistrationDone {
		println("Patient registration is done, go to next department")
	} else {
		fmt.Println("Reception registering patient")
		p.RegistrationDone = true
	}

	r.Next.Executer(p)
}
