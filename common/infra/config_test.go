package infra

import (
	"reflect"
	"testing"

	"github.com/zeromicro/go-queue/kq"
	"github.com/zeromicro/go-zero/core/discov"
)

func TestApplyEtcdKeepsServiceKey(t *testing.T) {
	c := Config{Etcd: EtcdConf{
		Hosts: []string{"etcd:2379"},
		User:  "user",
		Pass:  "pass",
	}}
	target := discov.EtcdConf{Key: "user.rpc"}

	c.ApplyEtcd(&target)

	if target.Key != "user.rpc" {
		t.Fatalf("service key changed to %q", target.Key)
	}
	if !reflect.DeepEqual(target.Hosts, c.Etcd.Hosts) {
		t.Fatalf("unexpected Etcd hosts: %v", target.Hosts)
	}
	if target.User != "user" || target.Pass != "pass" {
		t.Fatalf("Etcd credentials were not applied")
	}
}

func TestApplyKafkaKeepsServiceTopicAndGroup(t *testing.T) {
	c := Config{Kafka: KafkaConf{
		Brokers:  []string{"kafka:9092"},
		Username: "user",
		Password: "pass",
	}}
	target := kq.KqConf{
		Group: "email-consumer",
		Topic: "email",
	}

	c.ApplyKafka(&target)

	if target.Group != "email-consumer" || target.Topic != "email" {
		t.Fatalf("service Kafka semantics changed: group=%q topic=%q", target.Group, target.Topic)
	}
	if !reflect.DeepEqual(target.Brokers, c.Kafka.Brokers) {
		t.Fatalf("unexpected Kafka brokers: %v", target.Brokers)
	}
	if target.Username != "user" || target.Password != "pass" {
		t.Fatalf("Kafka credentials were not applied")
	}
}
