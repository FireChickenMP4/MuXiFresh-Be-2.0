package producer

import (
	"context"
	"strconv"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl/plain"
	"github.com/zeromicro/go-queue/kq"
)

type messageWriter interface {
	WriteMessages(ctx context.Context, messages ...kafka.Message) error
	Close() error
}

// Pusher writes user-auth messages to Kafka. Unlike kq.NewPusher, it applies
// the SASL credentials carried by kq.KqConf.
type Pusher struct {
	writer messageWriter
}

func NewPusher(c kq.KqConf) *Pusher {
	writer := &kafka.Writer{
		Addr:         kafka.TCP(c.Brokers...),
		Topic:        c.Topic,
		Balancer:     &kafka.LeastBytes{},
		Compression:  kafka.Snappy,
		RequiredAcks: kafka.RequireAll,
		WriteTimeout: 10 * time.Second,
		ReadTimeout:  10 * time.Second,
		MaxAttempts:  3,
		BatchTimeout: 10 * time.Millisecond,
		BatchSize:    1,
		Async:        false,
	}

	if c.Username != "" || c.Password != "" {
		writer.Transport = &kafka.Transport{
			SASL: plain.Mechanism{
				Username: c.Username,
				Password: c.Password,
			},
		}
	}

	return &Pusher{writer: writer}
}

func (p *Pusher) Push(ctx context.Context, value string) error {
	return p.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(strconv.FormatInt(time.Now().UnixNano(), 10)),
		Value: []byte(value),
	})
}

func (p *Pusher) Close() error {
	return p.writer.Close()
}
