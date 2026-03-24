package template

type IOtp interface {
	genRandomOTP(n int) string
	saveOTPCache(otp string)
	getMessage(otp string) string
	sendNotification(message string) error
}

type Otp struct {
	Iotp IOtp
}

func (o *Otp) GenAndSendOTP(n int) error {
	otp := o.Iotp.genRandomOTP(n)
	o.Iotp.saveOTPCache(otp)
	message := o.Iotp.getMessage(otp)
	err := o.Iotp.sendNotification(message)
	if err != nil {
		return err
	}
	return nil
}
