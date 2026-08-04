package producer

import (
	"context"
	"errors"
	"testing"

	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl/plain"
	"github.com/zeromicro/go-queue/kq"
)

type fakeWriter struct {
	messages []kafka.Message
	writeErr error
	closed   bool
}

func (w *fakeWriter) WriteMessages(_ context.Context, messages ...kafka.Message) error {
	w.messages = append(w.messages, messages...)
	return w.writeErr
}

func (w *fakeWriter) Close() error {
	w.closed = true
	return nil
}

func TestNewPusherConfiguresSASL(t *testing.T) {
	pusher := NewPusher(kq.KqConf{
		Brokers:  []string{"kafka.example.com:9094"},
		Topic:    "user-auth-email",
		Username: "root",
		Password: "secret",
	})

	writer, ok := pusher.writer.(*kafka.Writer)
	if !ok {
		t.Fatalf("unexpected writer type %T", pusher.writer)
	}

	transport, ok := writer.Transport.(*kafka.Transport)
	if !ok {
		t.Fatalf("unexpected transport type %T", writer.Transport)
	}

	mechanism, ok := transport.SASL.(plain.Mechanism)
	if !ok {
		t.Fatalf("unexpected SASL mechanism %T", transport.SASL)
	}
	if mechanism.Username != "root" || mechanism.Password != "secret" {
		t.Fatalf("unexpected credentials: username=%q password=%q", mechanism.Username, mechanism.Password)
	}
	if writer.Topic != "user-auth-email" {
		t.Fatalf("unexpected topic %q", writer.Topic)
	}
}

func TestNewPusherAllowsUnauthenticatedKafka(t *testing.T) {
	pusher := NewPusher(kq.KqConf{
		Brokers: []string{"kafka.example.com:9092"},
		Topic:   "user-auth-email",
	})

	writer := pusher.writer.(*kafka.Writer)
	if writer.Transport != nil {
		t.Fatalf("expected default transport, got %T", writer.Transport)
	}
}

func TestPusherPushAndClose(t *testing.T) {
	writeErr := errors.New("write failed")
	w := &fakeWriter{writeErr: writeErr}
	pusher := &Pusher{writer: w}

	err := pusher.Push(context.Background(), "payload")
	if !errors.Is(err, writeErr) {
		t.Fatalf("expected %v, got %v", writeErr, err)
	}
	if len(w.messages) != 1 {
		t.Fatalf("expected one message, got %d", len(w.messages))
	}
	if string(w.messages[0].Value) != "payload" {
		t.Fatalf("unexpected payload %q", w.messages[0].Value)
	}
	if len(w.messages[0].Key) == 0 {
		t.Fatal("expected a message key")
	}

	if err := pusher.Close(); err != nil {
		t.Fatalf("close failed: %v", err)
	}
	if !w.closed {
		t.Fatal("expected writer to be closed")
	}
}
