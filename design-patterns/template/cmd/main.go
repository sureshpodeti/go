package main

import (
	"designpatterns/template"
)

func main() {

	sms := &template.SMS{}
	// email := &template.Email{}

	otp := template.Otp{
		Iotp: sms,
	}
	otp.GenAndSendOTP(4)
}
