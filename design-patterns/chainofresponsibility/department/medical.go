package department

import (
	"designpatterns/chainofresponsibility/model"
	"fmt"
)

type Medical struct {
	Next Department
}

func (m *Medical) Executer(p *model.Patient) {
	if p.MedicineDone {
		fmt.Println("Medicine is already given to patient")
	} else {
		fmt.Println("Medical giving medicine to patient")
		p.MedicineDone = true
	}
	m.Next.Executer(p)
}
