package da

import (
	sendgrid "github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

// SendEmailSendgrid sends an email using SendGrid to the specified addresses.
func SendEmailSendgrid(apiKey string, fromName string, fromAddress string, targets []SendgridAddressConfig, message string) error {
	m := mail.NewV3Mail()
	m.Subject = "Status Monitor Alert"
	m.SetFrom(mail.NewEmail(fromName, fromAddress))

	p := mail.NewPersonalization()
	for _, info := range targets {
		to := mail.NewEmail(info.Name, info.Address)
		p.AddTos(to)
	}
	m.AddPersonalizations(p)

	m.AddContent(mail.NewContent("text/plain", message))

	request := sendgrid.GetRequest(apiKey, "/v3/mail/send", "https://api.sendgrid.com")
	request.Method = "POST"
	request.Body = mail.GetRequestBody(m)
	_, err := sendgrid.API(request)

	return err
}
