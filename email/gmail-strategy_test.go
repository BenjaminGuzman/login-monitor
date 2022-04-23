package email

import (
	"flag"
	"login-monitor/config"
	"testing"
)

var sender = flag.String("gmail-sender", "sender@example.com", "Sender email")
var recipient = flag.String("gmail-recipient", "recipient@example.com", "Recipient email")

var configFile = flag.String("gmail-config", "credentials.json", "Credentials file")
var tokenFile = flag.String("gmail-token", "token.json", "Token file")

func TestOAuth2(t *testing.T) {
	email := NewEmail(&GmailOAuth2Strategy{}).
		SetSender(config.NewEntity(*sender)).
		SetRecipient(config.NewEntity(*recipient)).
		SetSubject("Testing OAuth 2 strategy").
		SetHtmlMessage("<html><body><p>This is an <i>email</i> <b>test</b></p></body></html>").
		SetTextMessage("This is an email test (using text)").
		SetAttachments([]string{"./email-strategy.go"})
	_, err := email.InitStrategy(*configFile, *tokenFile)
	if err != nil {
		t.Error("Couldn't initiate Gmail OAuth2 strategy", err)
	}

	_, err = email.SendEmail()
	if err != nil {
		t.Error("Couldn't send email with Gmail OAuth2 strategy", err)
	} else {
		t.Log("Check your email!")
	}
}

func TestOAuth2PGP(t *testing.T) {
	email := NewEmail(&GmailOAuth2Strategy{}).
		SetSender(config.NewEntity(*sender)).
		SetRecipient(config.NewEntity(*recipient)).
		SetSubject("Testing OAuth 2 strategy").
		SetCc([]config.Entity{config.NewEntityWPGP("bg@benjaminguzman.dev", "0xE23BA39CD714EF8A")}).
		SetHtmlMessage("<html><body><p>This is an <i>email</i> <b>test</b></p></body></html>").
		SetTextMessage("This is an email test (using text)").
		SetAttachments([]string{"./email-strategy.go"})
	_, err := email.InitStrategy(*configFile, *tokenFile)
	if err != nil {
		t.Error("Couldn't initiate Gmail OAuth2 strategy", err)
	}

	_, err = email.SendPGPEmail()
	if err != nil {
		t.Error("Couldn't send email with Gmail OAuth2 strategy", err)
	} else {
		t.Log("Check your email!")
	}
}

func TestServiceAccount(t *testing.T) {
	return // Service account doesn't work yet
}
