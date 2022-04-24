# login monitor

**login monitor** is a script that allows you to monitor logins in a (*nix) remote machine.

This tool sends you email after any successful login\*.

The email can:

- be encrypted and signed
- contain attachments (e.g. logs)

\*You can modify this behaviour

See the full configuration tutorial in my 
[blog](https://medium.com/@GuzmanBenjamin/monitoring-logins-to-a-server-4cbd1515a4b9)

## Configuration

Check [schema.json](schema.json) and [config-example.json](config-example.json) to know more about the configuration.

## Go SMTP client

The code uses the [strategy](https://refactoring.guru/design-patterns/strategy) pattern, so it is easy to change
actual SMTP server (e.g. Gmail, postfix, Sendgrid...).

[gmail-strategy.go](email/gmail-strategy.go) is an implementation using 
[Gmail API](https://developers.google.com/gmail/api/quickstart/go)

[go-smtp-strategy.go](email/go-smtp-strategy.go) is an implementation using
Go's [`net/smtp`](https://pkg.go.dev/net/smtp) package

Note: in the code you'll find references to **pgp** and **gpg**. Because of their similarity these terms may end up
confusing you. So I'll clarify to you these terms briefly:

- **pgp** stands for Pretty Good Privacy. It's a software created by [Phil Zimmermann](https://philzimmermann.com), but
for sake of brevity I use the term _pgp_ to refer to [OpenPGP](https://www.openpgp.org/), which **is a standard**
- [**gpg**](https://www.openpgp.org/) stands for GNU Privacy Guard. It **is an implementation** of OpenPGP (pgp)

Therefore, if you see in my code something like `// create the pgp message` know that it may not be created with gpg 
but with other software like [rnpgp](https://www.rnpgp.org/) (used by thunderbird). Actually, it'd be nice if we could 
migrate the code to use something like 
[ProtonMail's OpenPGP implementation](https://pkg.go.dev/github.com/ProtonMail/go-crypto/openpgp) or
[GPGME](https://www.gnupg.org/software/gpgme/index.html)