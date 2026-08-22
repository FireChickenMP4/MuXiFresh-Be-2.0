package infra

import (
	"github.com/zeromicro/go-queue/kq"
	"github.com/zeromicro/go-zero/core/discov"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

// Config contains connection settings shared by services. It is loaded from a
// dedicated Nacos data ID so infrastructure changes do not require editing
// every service configuration.
type Config struct {
	MongoDB       MongoDBConf
	Redis         redis.RedisConf
	Etcd          EtcdConf
	Kafka         KafkaConf
	SMTP          SMTPConf
	ObjectStorage ObjectStorageConf
}

type MongoDBConf struct {
	URL string
	DB  string
}

// EtcdConf contains only the shared Etcd connection. Registry keys remain in
// each service configuration because they identify individual services.
type EtcdConf struct {
	Hosts              []string
	User               string
	Pass               string
	CertFile           string
	CertKeyFile        string
	CACertFile         string
	InsecureSkipVerify bool
}

type KafkaConf struct {
	Brokers  []string
	Username string
	Password string
}

type SMTPConf struct {
	Host     string
	Port     string
	UserName string
	Password string
}

type ObjectStorageConf struct {
	AccessKey  string
	SecretKey  string
	BucketName string
	DomainName string
}

func (c Config) ApplyEtcd(targets ...*discov.EtcdConf) {
	// go-zero's etcd client does not automatically use the User/Pass in config;
	// must call RegisterAccount explicitly, otherwise connecting to an
	// authenticated etcd fails with "user name is empty".
	if c.Etcd.User != "" && c.Etcd.Pass != "" {
		discov.RegisterAccount(c.Etcd.Hosts, c.Etcd.User, c.Etcd.Pass)
	}

	for _, target := range targets {
		if target == nil {
			continue
		}

		target.Hosts = append([]string(nil), c.Etcd.Hosts...)
		target.User = c.Etcd.User
		target.Pass = c.Etcd.Pass
		target.CertFile = c.Etcd.CertFile
		target.CertKeyFile = c.Etcd.CertKeyFile
		target.CACertFile = c.Etcd.CACertFile
		target.InsecureSkipVerify = c.Etcd.InsecureSkipVerify
	}
}

func (c Config) ApplyKafka(targets ...*kq.KqConf) {
	for _, target := range targets {
		if target == nil {
			continue
		}

		target.Brokers = append([]string(nil), c.Kafka.Brokers...)
		target.Username = c.Kafka.Username
		target.Password = c.Kafka.Password
	}
}
