package email

import (
	"flag"
	"login-monitor/config"
	"testing"
)

var sender1 = flag.String("go-smtp-sender", "sender@example.com", "Sender email")
var recipient1 = flag.String("go-smtp-recipient", "recipient@example.com", "Recipient email")

func TestSMTPStrategy(t *testing.T) {
	email := NewEmail(&GoSMTPStrategy{}).
		SetSender(config.NewEntity(*sender1)).
		SetRecipient(config.NewEntity(*recipient1)).
		SetSubject("Testing Go SMTP strategy").
		SetHtmlMessage("<html><body><p>This is an <i>email</i> <b>test</b></p></body></html>").
		SetTextMessage("This is an email test (using text)").
		SetAttachments([]string{"./email-strategy.go"})
	_, err := email.InitStrategy("", "", "", "127.0.0.1", "25") // requires postfix or similar installed
	if err != nil {
		t.Error("Couldn't initiate Go SMTP strategy", err)
	}

	_, err = email.SendEmail()
	if err != nil {
		t.Error("Couldn't send email with Go SMTP strategy", err)
	} else {
		t.Log("Check your email!")
	}
}

func TestSMTPStrategyPGP(t *testing.T) {
	email := NewEmail(&GmailOAuth2Strategy{}).
		SetSender(config.NewEntity(*sender1)).
		SetRecipient(config.NewEntity(*recipient1)).
		SetSubject("Testing Go SMTP strategy").
		SetCc([]config.Entity{config.NewEntityWPGP("bg@benjaminguzman.dev", "0xE23BA39CD714EF8A")}).
		SetHtmlMessage("<html><body><p>This is an <i>email</i> <b>test</b></p></body></html>").
		SetTextMessage("This is an email test (using text)").
		SetAttachments([]string{"./email-strategy.go"})
	_, err := email.InitStrategy("", "", "", "127.0.0.1", "25")
	if err != nil {
		t.Error("Couldn't initiate Go SMTP strategy", err)
	}

	_, err = email.SendPGPEmail()
	if err != nil {
		t.Error("Couldn't send email with Go SMTP strategy", err)
	} else {
		t.Log("Check your email!")
	}
}
