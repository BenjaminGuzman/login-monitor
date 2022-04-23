package config

type Entity struct {
	Email    string `json:"email"`    // email, e.g. sysadmin@example.com
	PGPKeyId string `json:"pgpKeyId"` // PGP key id, e.g. 0x7ADE4B572836C909 (it can be the email too, but just in some cases)
}

// NewEntity creates a new entity with Entity.PGPKeyId and Entity.Email equal to the given email
func NewEntity(email string) Entity {
	return Entity{Email: email, PGPKeyId: email}
}

// NewEntityWPGP creates a new entity with the given email and pgp key id
func NewEntityWPGP(email, pgpKeyId string) Entity {
	return Entity{Email: email, PGPKeyId: pgpKeyId}
}

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
