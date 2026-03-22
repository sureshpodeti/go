package main

import (
	"designpatterns/chainofresponsibility/department"
	"designpatterns/chainofresponsibility/model"
)

func main() {
	// Reception -> Doctor -> medical -> Cashier

	cashier := &department.Cashier{}
	medical := &department.Medical{Next: cashier}
	doctor := &department.Doctor{Next: medical}
	reception := &department.Reception{Next: doctor}

	patient := model.NewPatient(
		model.WithName("Tom"),
	)

	reception.Executer(patient)

}
