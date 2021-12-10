package api

import (
	"fmt"
	"os"
	"time"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
	gomail "gopkg.in/mail.v2"
)

var (
	gmail = os.Getenv("GMAIL_USER")
	pwd   = os.Getenv("GMAIL_PASSWORD")
)

type OTPAPIService struct {
	Key *otp.Key
}

func NewOTPService() *OTPAPIService {
	key, _ := totp.Generate(totp.GenerateOpts{
		Issuer:      "support@plantdoctor.com",
		AccountName: "support@plantdoctor.com",
		Period:      180,
	})
	return &OTPAPIService{Key: key}
}

func (o *OTPAPIService) GenerateCode() (string, error) {
	code, err := totp.GenerateCode(o.Key.Secret(), time.Now())
	if err != nil {
		return "", err
	}
	return code, nil
}

func (o *OTPAPIService) SendEmail(email string) (string, error) {
	dial := gomail.NewDialer("smtp.gmail.com", 587, gmail, pwd)
	res, err := dial.Dial()
	if err != nil {
		return "", err
	}
	code, err := o.GenerateCode()
	msg := gomail.NewMessage()
	msg.SetHeaders(map[string][]string{
		"From":    {msg.FormatAddress("support@plantdoctor.com", "PlantDoctor")},
		"To":      {email},
		"Subject": {"PlantDoctor 회원가입 인증코드"},
	})
	msg.SetBody("text/plain", fmt.Sprintf("인증코드: %s", code))
	if err := gomail.Send(res, msg); err != nil {
		return "", err
	}
	return code, nil
}

func (o *OTPAPIService) Validate(code string) bool {
	isValid := totp.Validate(code, o.Key.Secret())
	if isValid {
		return true
	} else {
		return false
	}
}
