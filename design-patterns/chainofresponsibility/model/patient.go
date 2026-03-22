package model

type Patient struct {
	Name                                                           string
	RegistrationDone, DoctorCheckupDone, MedicineDone, PaymentDone bool
}

func NewPatient(opts ...PatientOption) *Patient {
	patient := &Patient{
		RegistrationDone:  false,
		DoctorCheckupDone: false,
		MedicineDone:      false,
		PaymentDone:       false,
	}

	for _, opt := range opts {
		opt(patient)
	}
	return patient
}

type PatientOption func(*Patient)

func WithName(name string) PatientOption {
	return func(p *Patient) {
		p.Name = name
	}
}

func WithRegistrationDone(registrationDone bool) PatientOption {
	return func(p *Patient) {
		p.RegistrationDone = registrationDone
	}
}

func WithDoctorCheckupDone(doctorcheckupDone bool) PatientOption {
	return func(p *Patient) {
		p.DoctorCheckupDone = doctorcheckupDone
	}
}

func WithMedicineDone(medicineDone bool) PatientOption {
	return func(p *Patient) {
		p.MedicineDone = medicineDone
	}
}

func WithPaymentDone(payementDone bool) PatientOption {
	return func(p *Patient) {
		p.PaymentDone = payementDone
	}
}
