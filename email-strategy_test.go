package main

import (
	"log"
	"strings"
	"testing"
)

func getBoundary(payload string) string {
	boundaryStart := strings.Index(payload, "boundary=\"") + len("boundary=\"")
	boundaryEnd := boundaryStart + strings.IndexRune(payload[boundaryStart:], '"')
	return payload[boundaryStart:boundaryEnd]
}

func TestCreatePayload(t *testing.T) {
	email := NewEmail(nil).
		SetSender(NewEntity("bg@benjaminguzman.dev")).
		SetRecipient(NewEntity("benja@kobd.io")).
		SetSubject("Testing CreatePayload").
		SetHtmlMessage("<html><body><p>Testing <b>Create Payload</b></p></body></html>").
		SetTextMessage("Testing CreatePayload").
		SetAttachments([]string{"./download-audit-rules.sh"})

	actualPayload, err := email.CreatePayload()
	if err != nil {
		log.Fatal("Couldn't create payload", err)
	}

	outerBoundary := getBoundary(string(actualPayload))
	innerBoundary := getBoundary(string(actualPayload)[100:])

	expectedPayload := `Content-Type: multipart/mixed; boundary="outerboundary"
From: bg@benjaminguzman.dev
To: benja@kobd.io
Subject: Testing CreatePayload

This is a multi-part message in MIME format.
--outerboundary
Content-Type: multipart/alternative; boundary="innerboundary"

--innerboundary
Content-Transfer-Encoding: base64
Content-Type: text/plain; charset=utf-8; format=flowed; delsp=yes

VGVzdGluZyBDcmVhdGVQYXlsb2Fk

--innerboundary
Content-Transfer-Encoding: base64
Content-Type: text/html; charset=utf-8; format=flowed; delsp=yes

PGh0bWw+PGJvZHk+PHA+VGVzdGluZyA8Yj5DcmVhdGUgUGF5bG9hZDwvYj48L3A+PC9ib2R5Pjwv
aHRtbD4=

--innerboundary--
--outerboundary
Content-Disposition: attachment; filename="download-audit-rules.sh"
Content-Transfer-Encoding: base64
Content-Type: text/plain; charset=utf-8; name="download-audit-rules.sh"

IyEvYmluL2Jhc2gKCmlmICEgY2QgL2V0Yy9hdWRpdC9ydWxlcy5kOyB0aGVuCiAgICBlY2hvICJS
dWxlcyBzaG91bGQgYmUgZG93bmxvYWRlZCBpbiAvZXRjL2F1ZGl0L3J1bGVzLmQiCiAgICBlY2hv
IC1lICJcMDMzWzkzbVJ1biB0aGlzIHNjcmlwdCB3aXRoIHJvb3QgcHJpdmlsZWdlc1wwMzNbMG0i
CiAgICBleGl0IDEKZmkKCmRlY2xhcmUgLUEgdXJscyAjIGFzc29jaWF0aXZlIGFycmF5IG9mIHVy
bHMKdXJscz0oCiAgICBbImxvZ2ludWlkIGlubXV0YWJsZSJdPSJodHRwczovL3Jhdy5naXRodWJ1
c2VyY29udGVudC5jb20vbGludXgtYXVkaXQvYXVkaXQtdXNlcnNwYWNlL21hc3Rlci9ydWxlcy8x
MS1sb2dpbnVpZC5ydWxlcyIKICAgIFsiaW5zdGFsbGVycyJdPSJodHRwczovL3Jhdy5naXRodWJ1
c2VyY29udGVudC5jb20vbGludXgtYXVkaXQvYXVkaXQtdXNlcnNwYWNlL21hc3Rlci9ydWxlcy80
NC1pbnN0YWxsZXJzLnJ1bGVzIgogICAgWyJwY2ktZHNzIHYzLjEiXT0iaHR0cHM6Ly9yYXcuZ2l0
aHVidXNlcmNvbnRlbnQuY29tL2xpbnV4LWF1ZGl0L2F1ZGl0LXVzZXJzcGFjZS9tYXN0ZXIvcnVs
ZXMvMzAtcGNpLWRzcy12MzEucnVsZXMiCiAgICBbIklubXV0YWJsZSBjb25maWd1cmF0aW9uIl09
Imh0dHBzOi8vcmF3LmdpdGh1YnVzZXJjb250ZW50LmNvbS9saW51eC1hdWRpdC9hdWRpdC11c2Vy
c3BhY2UvbWFzdGVyL3J1bGVzLzk5LWZpbmFsaXplLnJ1bGVzIgopCgpmb3IgcnVsZU5hbWUgaW4g
IiR7IXVybHNbQF19IjsgZG8KICAgIGVjaG8gLWUgIkRvd25sb2FkaW5nIHJ1bGUgXDAzM1s5Mm0k
cnVsZU5hbWVcMDMzWzBtLi4uIgogICAgY3VybCAtTyAiJHt1cmxzWyRydWxlTmFtZV19IiAtLXBy
b2dyZXNzLWJhcgpkb25lCgplY2hvICJEb25lLiIKCmVjaG8gLWUgIlxuTm93IGlzIHRpbWUgdG8g
dXBkYXRlIHJ1bGVzIGNvbmZpZ3VyYXRpb24gKGluc2lkZSAvZXRjL2F1ZGl0L3J1bGVzLmQpIGFz
IG5lZWRlZC4iCmVjaG8gLWUgIlwwMzNbOTdtUmVib290XDAzM1swbSB3aGVuIHlvdSBmaW5pc2gg
dG8gYXBwbHkgY2hhbmdlcyIK

--outerboundary--
`
	expectedPayload = strings.ReplaceAll(expectedPayload, "outerboundary", outerBoundary)
	expectedPayload = strings.ReplaceAll(expectedPayload, "innerboundary", innerBoundary)
	expectedPayload = strings.ReplaceAll(expectedPayload, "\n", "\r\n")

	if expectedPayload != string(actualPayload) {
		t.Errorf("Payloads differ. Expected: %s, Actual: %s", expectedPayload, string(actualPayload))
	}
}
