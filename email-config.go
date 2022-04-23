package main

type EmailConfig struct {
	Sender         Entity   `json:"sender"`
	FakeSender     string   `json:"fakeSender"`
	Recipient      Entity   `json:"recipient"`
	Cc             []Entity `json:"cc"`
	Subject        string   `json:"subject"`
	TextMessage    string   `json:"textMessage"`
	HTMLMessage    string   `json:"htmlMessage"`
	Attachments    []string `json:"attachments"`
	SenderPassFile string   `json:"senderPassFile"` // path to the sender's private key passphrase (required if the message is signed)
}
