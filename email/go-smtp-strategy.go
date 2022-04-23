package email

import (
	"bytes"
	"fmt"
	"net/smtp"
	"strings"
)

type GoSMTPStrategy struct {
	auth    smtp.Auth
	address string
}

// Init initiates a new gmail.Service (required by other methods)
// 1st param: identity
// 2nd param: username
// 3rd param: password
// 4th param: host
// 5th param: port
// See smtp.PlainAuth for more information about the parameters
// Returns nothing
func (s *GoSMTPStrategy) Init(params ...interface{}) (interface{}, error) {
	identity := fmt.Sprint(params[0])
	username := fmt.Sprint(params[1])
	password := fmt.Sprint(params[2])
	host := fmt.Sprint(params[3])
	port := fmt.Sprint(params[4])

	s.auth = smtp.PlainAuth(identity, username, password, host)
	s.address = host + ":" + port

	return nil, nil
}

// SendEmail sends the email with the gmail api. Returns nothing but an error, if any.
func (s *GoSMTPStrategy) SendEmail(payload []byte, sender string) (interface{}, error) {
	to := extractCc(payload)
	to = append(to, extractRecipient(payload))

	err := smtp.SendMail(s.address, s.auth, sender, to, payload)
	if err != nil {
		if strings.Contains(err.Error(), "smtp: server doesn't support AUTH") {
			return s.sendMailNoAuth(payload, sender, to...)
		}
		return nil, err
	}
	return nil, nil
}

func (s *GoSMTPStrategy) sendMailNoAuth(payload []byte, sender string, recipients ...string) (interface{}, error) {
	client, err := smtp.Dial(s.address)
	if err != nil {
		return nil, fmt.Errorf("couldn't connect to %s. %w", s.address, err)
	}
	defer client.Close()

	if err = client.Mail(sender); err != nil {
		return nil, err
	}
	for _, rcpt := range recipients {
		if err = client.Rcpt(rcpt); err != nil {
			return nil, err
		}
	}

	w, err := client.Data()
	if err != nil {
		return nil, err
	}

	_, err = w.Write(payload)
	if err != nil {
		return nil, err
	}

	if err = w.Close(); err != nil {
		return nil, err
	}
	err = client.Quit()
	return nil, err
}

func extractRecipient(payload []byte) string {
	rcptStart := bytes.Index(payload, []byte("To:"))
	if rcptStart == -1 {
		return ""
	}
	rcptStart += len([]byte("To:"))

	var rcptEnd int
	for i := rcptStart; i < len(payload); i++ {
		if payload[i] == '\n' {
			rcptEnd = i + 1
			break
		}
	}

	recipient := strings.TrimSpace(string(payload[rcptStart:rcptEnd]))
	return recipient
}

func extractCc(payload []byte) []string {
	ccStart := bytes.Index(payload, []byte("Cc:"))
	if ccStart == -1 {
		return []string{}
	}
	ccStart += len([]byte("Cc:"))

	var ccEnd int
	for i := ccStart; i < len(payload); i++ {
		if payload[i] == '\n' {
			ccEnd = i + 1
			break
		}
	}

	recipients := strings.Split(strings.TrimSpace(string(payload[ccStart:ccEnd])), ",")
	for i, recipient := range recipients {
		recipients[i] = strings.TrimSpace(recipient)
	}
	return recipients
}
