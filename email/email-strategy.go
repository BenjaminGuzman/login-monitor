package email

import (
	"bytes"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"
	"io/fs"
	"login-monitor/config"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// MaxLen Max line length for the email
const MaxLen = 76

type Email struct {
	sender         config.Entity
	fakeSender     string
	recipient      config.Entity
	cc             []config.Entity
	subject        string
	textMessage    string
	htmlMessage    string
	attachments    []string
	senderPassFile string // path to the sender's private key passphrase (required if the message is signed)

	initiated bool
	strategy  EmailStrategy
}

// NewEmail creates a new Email with the given strategy
func NewEmail(strategy EmailStrategy) *Email {
	return &Email{
		cc:          []config.Entity{},
		attachments: []string{},
		strategy:    strategy,
	}
}

// Init (re) initiates the email object by calling setters. Particularly useful if email object was created by deserialization
func (e *Email) Init() *Email {
	// these setters have specific logic
	return e.SetSubject(e.subject).
		SetTextMessage(e.textMessage).
		SetHtmlMessage(e.htmlMessage)
}

func (e *Email) InitFromConfig(c *config.EmailConfig) *Email {
	return e.SetSubject(c.Subject).
		SetCc(c.Cc).
		SetSender(c.Sender).
		SetFakeSender(c.FakeSender).
		SetAttachments(c.Attachments).
		SetRecipient(c.Recipient).
		SetHtmlMessage(c.HTMLMessage).
		SetTextMessage(c.TextMessage).
		SetSenderPassFile(c.SenderPassFile)
}

func (e *Email) Sender() config.Entity {
	return e.sender
}

func (e *Email) FakeSender() string {
	if e.fakeSender == "" {
		return e.sender.Email
	}
	return e.fakeSender
}

func (e *Email) SenderPassFile() string {
	return e.senderPassFile
}

func (e *Email) Recipient() config.Entity {
	return e.recipient
}

func (e *Email) Cc() []config.Entity {
	return e.cc
}

func (e *Email) Subject() string {
	return e.subject
}

func (e *Email) TextMessage() string {
	return e.textMessage
}

func (e *Email) HtmlMessage() string {
	return e.htmlMessage
}

func (e *Email) Attachments() []string {
	return e.attachments
}

func (e *Email) SetSender(sender config.Entity) *Email {
	e.sender = sender
	return e
}

func (e *Email) SetFakeSender(fakeSender string) *Email {
	e.fakeSender = fakeSender
	return e
}

func (e *Email) SetSenderPassFile(passFile string) *Email {
	e.senderPassFile = passFile
	return e
}

func (e *Email) SetRecipient(recipient config.Entity) *Email {
	e.recipient = recipient
	return e
}

func (e *Email) SetCc(cc []config.Entity) *Email {
	e.cc = cc
	return e
}

func (e *Email) SetSubject(subject string) *Email {
	e.subject = ReplacePlaceholders(subject)
	return e
}

func (e *Email) SetTextMessage(textMessage string) *Email {
	if trimmed := strings.TrimSpace(textMessage); len(trimmed) > 5 && trimmed[len(trimmed)-4:] == ".txt" { // content may be a file
		if contents, err := os.ReadFile(textMessage); err == nil {
			textMessage = string(contents)
		}
	}
	e.textMessage = ReplacePlaceholders(textMessage)
	return e
}

func (e *Email) SetHtmlMessage(htmlMessage string) *Email {
	if trimmed := strings.TrimSpace(htmlMessage); len(trimmed) > 6 && trimmed[len(trimmed)-5:] == ".html" { // content may be a file
		if contents, err := os.ReadFile(htmlMessage); err == nil {
			htmlMessage = string(contents)
		}
	}

	e.htmlMessage = ReplacePlaceholders(htmlMessage)
	return e
}

func (e *Email) SetAttachments(attachments []string) *Email {
	realAttachments := make([]string, 0, len(attachments))
	for _, attachment := range attachments {
		stat, err := os.Stat(attachment)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Error with attachment '%s'. Ignoring it. %s", attachment, err)
			continue
		}
		if !stat.IsDir() {
			realAttachments = append(realAttachments, attachment)
			continue
		}

		// add all files inside the directory
		log.Debugf("Attachment %s is a directory, walking the directory", attachment)
		_ = filepath.WalkDir(attachment, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				_, _ = fmt.Fprintf(os.Stderr, "Error with attachment '%s'. Ignoring it. %s", path, err)
				return nil // just ignore the error and continue, but don't add the file
			}
			if d.IsDir() {
				return nil // walk into that directory, but obviously don't add it as an attachment
			}

			realAttachments = append(realAttachments, path)
			return nil
		})
	}
	e.attachments = realAttachments
	return e
}

// CCEmails Returns only the emails in Email.cc
func (e *Email) CCEmails() []string {
	if e.cc == nil {
		return nil
	}

	emails := make([]string, 0, len(e.cc))
	for _, entity := range e.cc {
		if entity.Email != "" {
			emails = append(emails, entity.Email)
		}
	}
	return emails
}

// CCPGPKeyIds Returns the values of Email.cc
func (e *Email) CCPGPKeyIds() []string {
	if e.cc == nil {
		return nil
	}

	keyIds := make([]string, 0, len(e.cc))
	for _, entity := range e.cc {
		if entity.PGPKeyId != "" {
			keyIds = append(keyIds, entity.PGPKeyId)
		}
	}
	return keyIds
}

type EmailStrategy interface {
	// Init initialize the strategy. Read config files, credentials, generate tokens, etc..
	Init(...interface{}) (interface{}, error)

	// SendEmail sends the given payload as email with the specified sender
	// (it is recommended to be the same as provided to CreatePayload, but it's not necessary)
	SendEmail(payload []byte, sender string) (interface{}, error)
}

// InitStrategy simply calls EmailStrategy.Init on the context's strategy
func (e *Email) InitStrategy(params ...interface{}) (interface{}, error) {
	if r, err := e.strategy.Init(params...); err != nil {
		return r, err
	} else {
		e.initiated = true
		return r, err
	}
}

// SendEmail Sends the email. Prior to calling this method (or any other method on e) you should set fields via setters
func (e *Email) SendEmail() (interface{}, error) {
	if !e.initiated {
		return nil, errors.New("strategy needs to be initiated")
	}

	log.Debugln("Creating email payload")
	payload, err := e.CreatePayload()
	if err != nil {
		return nil, err
	}
	log.Debugln("Done creating email payload")

	log.Debugln("Sending email payload")
	res, err := e.strategy.SendEmail(payload, e.Sender().Email)
	if err != nil {
		return nil, err
	}
	log.Debugln("Done sending email payload")

	return res, nil
}

// SendPGPEmail Sends a PGP-encrypted email using the context's strategy.
// Prior to calling this method (or any other method on e) you should set fields via setters
//
// See also Email.CreatePGPPayload
func (e *Email) SendPGPEmail() (interface{}, error) {
	if !e.initiated {
		return nil, errors.New("strategy needs to be initiated")
	}

	payload, err := e.CreatePGPPayload()
	if err != nil {
		return nil, err
	}

	res, err := e.strategy.SendEmail(payload, e.Sender().Email)
	if err != nil {
		return nil, err
	}

	return res, nil
}

// IsPGPCandidate tells if the email can be a PGP email. It is considered a candidate if at least one of the recipients'
// (Email.Recipient or Email.Cc) public key is present in the GPG keyring
func (e *Email) IsPGPCandidate() bool {
	recipientsKeyIds := append(e.CCPGPKeyIds(), e.Recipient().PGPKeyId) // This may seem wrong, but is actually right because we modify a copy of the Cc emails (getter returns such copy)
	return recipientsKeyExist(true, recipientsKeyIds...)
}

func createBasicHeaders(from, to, subject string) string {
	return fmt.Sprintf(
		"From: %s\r\n"+
			"To: %s\r\n"+
			"Subject: %s\r\n",
		from,
		to,
		subject,
	)
}

func (e *Email) createCCHeader() string {
	strBuilder := strings.Builder{}
	if ccEmails := e.CCEmails(); ccEmails != nil && len(ccEmails) > 0 {
		strBuilder.WriteString("Cc: ")
		strBuilder.WriteString(strings.Join(ccEmails, ","))
		strBuilder.WriteString("\r\n")
	}
	return strBuilder.String()
}

// CreateMessagePayload creates a multipart/alternative payload with the text plain and html message specified in e
//
// This is a pure function, i.e. e is not modified
func (e *Email) CreateMessagePayload() ([]byte, error) {
	textBytes := []byte(e.TextMessage())

	payload := bytes.Buffer{}
	mpWriter := multipart.NewWriter(&payload)

	payload.Grow(len(textBytes))
	payload.WriteString(fmt.Sprintf("Content-Type: multipart/alternative; boundary=\"%s\"\r\n\r\n", mpWriter.Boundary()))

	// write text/plain
	textPlain, err := mpWriter.CreatePart(textproto.MIMEHeader{
		"Content-Type":              {fmt.Sprintf("%s; format=flowed; delsp=yes", http.DetectContentType(textBytes))},
		"Content-Transfer-Encoding": {"base64"},
	})
	if err != nil {
		return nil, err
	}
	if _, err = textPlain.Write(Wrap(Base64Encode(textBytes), MaxLen, "\r\n")); err != nil {
		return nil, err
	}

	// write text/html
	if e.htmlMessage != "" {
		htmlBytes := []byte(e.HtmlMessage())
		// https://stackoverflow.com/questions/25710599/content-transfer-encoding-7bit-or-8-bit
		// this is preferred to be a quoted-printable encoding, but for simplicity of the code let's just leave it as base64
		textHTML, err := mpWriter.CreatePart(textproto.MIMEHeader{
			"Content-Type":              {fmt.Sprintf("%s; format=flowed; delsp=yes", http.DetectContentType(htmlBytes))},
			"Content-Transfer-Encoding": {"base64"},
		})

		if err != nil {
			return nil, err
		}
		if _, err = textHTML.Write(Wrap(Base64Encode(htmlBytes), MaxLen, "\r\n")); err != nil {
			return nil, err
		}
	}

	_ = mpWriter.Close()

	return payload.Bytes(), nil
}

// CreatePayload creates a multipart payload with all the data specified in e
//
// This is a pure function, i.e. e is not modified
func (e *Email) CreatePayload() ([]byte, error) {
	payload := bytes.Buffer{}
	mpWriter := multipart.NewWriter(&payload)

	// write email headers
	payload.WriteString(fmt.Sprintf("Content-Type: multipart/mixed; boundary=\"%s\"\r\n", mpWriter.Boundary()))
	payload.WriteString(createBasicHeaders(e.FakeSender(), e.Recipient().Email, e.subject))

	// write CC headers
	payload.WriteString(e.createCCHeader())
	payload.WriteString("\r\nThis is a multi-part message in MIME format.\r\n")

	// write message
	msgPayload, err := e.CreateMessagePayload()
	if err != nil {
		return nil, err
	}
	// yes, write directly on the payload. Writing on a new part would add unnecessary linebreaks
	payload.WriteString(fmt.Sprintf("--%s\r\n", mpWriter.Boundary()))
	payload.Write(msgPayload)

	// write files
	for _, attachmentPath := range e.attachments {
		file, err := os.Open(attachmentPath)
		if err != nil {
			return nil, err
		}

		fileBytes, err := io.ReadAll(file)
		_ = file.Close()
		if err != nil {
			return nil, err
		}

		fileContentType := http.DetectContentType(fileBytes)
		fileName := filepath.Base(attachmentPath)

		// write file headers
		filePart, err := mpWriter.CreatePart(textproto.MIMEHeader{
			"Content-Type":              {fmt.Sprintf("%s; name=\"%s\"", fileContentType, fileName)},
			"Content-Disposition":       {fmt.Sprintf("attachment; filename=\"%s\"", fileName)},
			"Content-Transfer-Encoding": {"base64"},
		})
		if err != nil {
			return nil, err
		}

		// write file bytes
		if _, err := filePart.Write(Wrap(Base64Encode(fileBytes), MaxLen, "\r\n")); err != nil {
			return nil, err
		}
	}

	_ = mpWriter.Close()

	return payload.Bytes(), nil
}

// CreatePGPPayload Similarly to Email.CreatePayload, this creates a multipart payload encrypted with the recipient's public key.
// IMPORTANT: This requires gpg installed on the system
//
// Recipient's public key must exists in the gpg keyring, otherwise an error is returned.
//
// If sender private key is in the keyring, the message will also be signed
// It is recommended that sender's public key is present in the keyring. See below.
//
// Some clients (thunderbird) don't handle the single recipient case (when sender = recipient) so it may show some error.
// But, that's actually not right. Info about multiple recipient case:
// https://security.stackexchange.com/questions/8245/gpg-file-size-with-multiple-recipients
func (e *Email) CreatePGPPayload() ([]byte, error) {
	payload := bytes.Buffer{}
	mpWriter := multipart.NewWriter(&payload)

	// write email headers
	payload.WriteString(fmt.Sprintf("Content-Type: multipart/encrypted; protocol=\"application/pgp-encrypted\"; boundary=\"%s\"\r\n", mpWriter.Boundary()))
	payload.WriteString(createBasicHeaders(e.FakeSender(), e.Recipient().Email, e.subject))

	// write CC headers
	payload.WriteString(e.createCCHeader())
	payload.WriteString("\r\nThis is an OpenPGP/MIME encrypted message (RFC 4880 and 3156)\r\n")

	// write PGP header for encrypted message and PGP version
	part, err := mpWriter.CreatePart(textproto.MIMEHeader{
		"Content-Type":        {"application/pgp-encrypted"},
		"Content-Description": {"PGP/MIME version identification"},
	})
	if err != nil {
		return nil, err
	}
	if _, err = part.Write([]byte("Version: 1\r\n")); err != nil {
		return nil, err
	}

	// create plain text body (the plain payload inside the encrypted payload)
	body, err := e.CreatePayload()
	if err != nil {
		return nil, err
	}

	// sign body and create a wrapper consisting of 2 parts: body and signature
	senderPrivKeyExists := recipientsKeyExist(false, e.Sender().PGPKeyId)
	if senderPrivKeyExists {
		var signature []byte
		if signature, err = pgpSign(body, e.Sender().PGPKeyId, e.senderPassFile); err != nil {
			return nil, err
		}
		wrapper := bytes.Buffer{}
		wrapperWriter := multipart.NewWriter(&wrapper)
		wrapper.WriteString(fmt.Sprintf(
			"Content-Type: multipart/signed; "+
				"micalg=pgp-sha256; "+ // TODO change MICalg
				"protocol=\"application/pgp-signature\"; "+
				"boundary=\"%s\"",
			wrapperWriter.Boundary(),
		))
		wrapper.WriteString("\r\n\r\nThis is an OpenPGP/MIME signed message (RFC 4880 and 3156)\r\n")

		// write body (write it without creating a new part because the body itself contains all the required headers)
		wrapper.WriteString(fmt.Sprintf("--%s\r\n%s\r\n", wrapperWriter.Boundary(), body))

		// write signature
		sigPart, err := wrapperWriter.CreatePart(textproto.MIMEHeader{
			"Content-Type":        {"application/pgp-signature; name=\"OpenPGP_signature.asc\""},
			"Content-Description": {"OpenPGP digital signature"},
			"Content-Disposition": {"attachment; filename=\"OpenPGP_signature\""},
		})
		if err != nil {
			return nil, err
		}
		if _, err = sigPart.Write(signature); err != nil {
			return nil, err
		}
		_ = wrapperWriter.Close()

		body = wrapper.Bytes() // new body = (previous) body + signature
	}

	// encrypt plain text body
	var encryptedBody []byte
	recipientsKeyIds := append(e.CCPGPKeyIds(), e.Recipient().PGPKeyId)       // This may seem wrong, but is actually right because we modify a copy of the Cc emails (getter returns such copy)
	if senderPrivKeyExists || recipientsKeyExist(true, e.Sender().PGPKeyId) { // If private key exist, public key must exist (or at least can be obtained from the private key)
		encryptedBody, err = pgpEncrypt(body, e.Sender().PGPKeyId, recipientsKeyIds...) // encrypt for both
	} else {
		encryptedBody, err = pgpEncrypt(body, "", recipientsKeyIds...) // encrypt only for recipient as sender key doesn't exist
	}
	if err != nil {
		return nil, err
	}

	// write encrypted body
	part, err = mpWriter.CreatePart(textproto.MIMEHeader{
		"Content-Type":        {"application/octet-stream; name=\"encrypted.asc\""},
		"Content-Description": {"OpenPGP encrypted message"},
		"Content-Disposition": {"inline; filename=\"encrypted.asc\""},
	})
	if err != nil {
		return nil, err
	}
	if _, err = part.Write(encryptedBody); err != nil {
		return nil, err
	}

	_ = mpWriter.Close()

	return payload.Bytes(), nil
}

// TODO migrate functions below to GPGME or https://pkg.go.dev/github.com/ProtonMail/go-crypto or github.com/ProtonMail/gopenpgp/v2

// pgpEncrypt Encrypt the given data using gpg and the public keys for the given recipients
//
// If senderId is not empty and is associated to a public key, message will be encrypted for the sender too
//
// If senderId is not empty and is associated to a private key, message will be signed
func pgpEncrypt(data []byte, senderId string, recipientsIds ...string) ([]byte, error) {
	var initialArgs []string
	if senderId != "" {
		// encrypt the message for the sender too
		initialArgs = []string{"--batch", "--pinentry-mode", "loopback", "--encrypt", "--armor", "--trust-model", "always", "--recipient", senderId}
	} else {
		initialArgs = []string{"--batch", "--pinentry-mode", "loopback", "--encrypt", "--armor", "--trust-model", "always"}
	}
	gpgArgs := make([]string, len(initialArgs)+len(recipientsIds)*2)
	copy(gpgArgs, initialArgs)
	for i := len(initialArgs); i < len(gpgArgs); i += 2 {
		gpgArgs[i] = "--recipient"
		gpgArgs[i+1] = recipientsIds[(i-len(initialArgs))/2]
	}

	log.Debugln("Encrypting data. Executing gpg", gpgArgs)
	encryptCmd := exec.Command("gpg", gpgArgs...)
	var stdout, stderr bytes.Buffer
	encryptCmd.Stderr = &stderr
	encryptCmd.Stdout = &stdout

	stdin, err := encryptCmd.StdinPipe()
	if err != nil {
		return nil, err
	}

	go func() {
		defer stdin.Close() // we need to close it, otherwise gpg will keep reading from it and block the thread
		_, err = stdin.Write(data)
	}()

	if err := encryptCmd.Start(); err != nil {
		return nil, err
	}

	if err := encryptCmd.Wait(); err != nil { // if recipient's key doesn't exist, this will return an error
		return nil, fmt.Errorf("error while encrypting PGP message, pgp stderr: \"%s\". %w", stderr.Bytes(), err)
	}

	encrypted := stdout.Bytes()
	return encrypted, nil
}

// pgpSign signs the given data and returns the signature
func pgpSign(data []byte, senderId string, passphraseFile string) ([]byte, error) {
	gpgArgs := []string{
		"--batch",
		"--pinentry-mode", "loopback",
		"--armor",
		"--trust-model", "always",
		"--detach-sig",
		"--local-user", senderId,
	}
	if passphraseFile != "" {
		gpgArgs = append(gpgArgs, "--passphrase-file", passphraseFile)
	}

	log.Debugln("Signing data. Executing gpg", gpgArgs)
	signCmd := exec.Command("gpg", gpgArgs...)
	var stdout, stderr bytes.Buffer
	signCmd.Stderr = &stderr
	signCmd.Stdout = &stdout

	stdin, err := signCmd.StdinPipe()
	if err != nil {
		return nil, err
	}

	go func() {
		defer stdin.Close() // we need to close it, otherwise gpg will keep reading from it and block the thread
		_, err = stdin.Write(data)
	}()

	if err := signCmd.Start(); err != nil {
		return nil, err
	}

	if err := signCmd.Wait(); err != nil { // if recipient's key doesn't exist, this will return an error
		return nil, fmt.Errorf("error while signing PGP message, pgp stderr: \"%s\". %w", stderr.Bytes(), err)
	}

	signed := stdout.Bytes()
	return signed, nil
}

// recipientsKeyExist Tells whether the public/private exists in the gpg keyring for ANY of the recipients given.
func recipientsKeyExist(public bool, recipients ...string) bool {
	gpgArgs := []string{"--list-public-keys", "--batch", "--with-colons"} // with-colons is actually not needed (for now)
	if !public {
		gpgArgs[0] = "--list-secret-keys"
	}
	gpgArgs = append(gpgArgs, recipients...)

	if err := exec.Command("gpg", gpgArgs...).Run(); err != nil { // recipient's key doesn't exist
		return false
	}
	return true
}
