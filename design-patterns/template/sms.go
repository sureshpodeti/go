package template

import "fmt"

type SMS struct{}

func (s *SMS) genRandomOTP(n int) string {
	randomOTP := "1234"
	fmt.Printf("SMS: generating random otp %s\n", randomOTP)
	return randomOTP
}

func (s *SMS) saveOTPCache(otp string) {
	fmt.Printf("SMS: saving otp: %s to cache\n", otp)
}

func (s *SMS) getMessage(otp string) string {
	return "SMS OTP for login is " + otp
}

func (s *SMS) sendNotification(message string) error {
	fmt.Printf("SMS: sending sms: %s\n", message)
	return nil
}
