package department

import (
	"designpatterns/chainofresponsibility/model"
	"fmt"
)

type Doctor struct {
	Next Department
}

func (d *Doctor) Executer(p *model.Patient) {
	if p.DoctorCheckupDone {
		fmt.Println("Doctor checkup is already done!")
	} else {
		fmt.Println("Doctor Checking patient")
		p.DoctorCheckupDone = true
	}

	d.Next.Executer(p)
}
