package email

import (
	"crypto/tls"
	"errors"
	"net/smtp"
	"testing"

	jordanemail "github.com/jordan-wright/email"
)

func TestSendUsesImplicitTLS(t *testing.T) {
	previousInfo := eInfo
	previousSendWithTLS := sendWithTLS
	t.Cleanup(func() {
		eInfo = previousInfo
		sendWithTLS = previousSendWithTLS
	})

	eInfo = EmailInfo{
		Host:     "smtp.example.com",
		Port:     "465",
		UserName: "sender@example.com",
		Password: "secret",
	}

	expectedError := errors.New("stop before redis")
	var capturedAddress string
	var capturedConfig *tls.Config
	sendWithTLS = func(message *jordanemail.Email, address string, auth smtp.Auth, config *tls.Config) error {
		if message.From != eInfo.UserName {
			t.Fatalf("unexpected sender: %q", message.From)
		}
		if len(message.To) != 1 || message.To[0] != "recipient@example.com" {
			t.Fatalf("unexpected recipients: %v", message.To)
		}
		if auth == nil {
			t.Fatal("SMTP auth must be configured")
		}
		capturedAddress = address
		capturedConfig = config
		return expectedError
	}

	err := Send("recipient@example.com", "set_password")
	if !errors.Is(err, expectedError) {
		t.Fatalf("expected send error %v, got %v", expectedError, err)
	}
	if capturedAddress != "smtp.example.com:465" {
		t.Fatalf("unexpected SMTP address: %q", capturedAddress)
	}
	if capturedConfig == nil {
		t.Fatal("TLS config must be provided")
	}
	if capturedConfig.ServerName != "smtp.example.com" {
		t.Fatalf("unexpected TLS server name: %q", capturedConfig.ServerName)
	}
	if capturedConfig.MinVersion != tls.VersionTLS12 {
		t.Fatalf("unexpected minimum TLS version: %d", capturedConfig.MinVersion)
	}
	if capturedConfig.InsecureSkipVerify {
		t.Fatal("TLS certificate verification must remain enabled")
	}
}
