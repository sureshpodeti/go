package template

import "fmt"

type Email struct{}

func (e *Email) genRandomOTP(n int) string {
	randomOTP := "1234"
	fmt.Printf("SMS: generating random otp %s\n", randomOTP)
	return randomOTP
}

func (e *Email) saveOTPCache(otp string) {
	fmt.Printf("SMS: saving otp: %s to cache\n", otp)
}

func (e *Email) getMessage(otp string) string {
	return "SMS OTP for login is " + otp
}

func (e *Email) sendNotification(message string) error {
	fmt.Printf("SMS: sending sms: %s\n", message)
	return nil
}
