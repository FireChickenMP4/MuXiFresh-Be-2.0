package nacos

import (
	"fmt"
	"os"

	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/nacos-group/nacos-sdk-go/vo"
	"sigs.k8s.io/yaml"
)

const (
	defaultGroup       = "PROD"
	defaultInfraDataID = "infra"
)

type LoadOption struct {
	Group  string
	DataId string
	Target any
}

// Service returns the Nacos load option for a service-owned configuration.
// NACOS_GROUP can override the default group for all services in an
// environment.
func Service(dataID string, target any) LoadOption {
	return LoadOption{
		Group:  valueOrDefault("NACOS_GROUP", defaultGroup),
		DataId: dataID,
		Target: target,
	}
}

// Infra returns the Nacos load option for shared infrastructure. By default it
// uses the service group and the "infra" data ID. Both can be overridden for
// deployments that keep infrastructure in a separate group or data ID.
func Infra(target any) LoadOption {
	group := os.Getenv("NACOS_INFRA_GROUP")
	if group == "" {
		group = valueOrDefault("NACOS_GROUP", defaultGroup)
	}

	return LoadOption{
		Group:  group,
		DataId: valueOrDefault("NACOS_INFRA_DATA_ID", defaultInfraDataID),
		Target: target,
	}
}

func Load(options ...LoadOption) error {
	client, err := newNacosClient()
	if err != nil {
		return err
	}

	for _, opt := range options {
		if opt.Group == "" {
			return fmt.Errorf("nacos group must not be empty")
		}
		if opt.DataId == "" {
			return fmt.Errorf("nacos data ID must not be empty")
		}
		if opt.Target == nil {
			return fmt.Errorf("target for %s must not be nil", opt.DataId)
		}

		content, err := client.GetConfig(vo.ConfigParam{
			DataId: opt.DataId,
			Group:  opt.Group,
		})
		if err != nil {
			return fmt.Errorf("load %s failed: %w", opt.DataId, err)
		}

		if err := yaml.Unmarshal([]byte(content), opt.Target); err != nil {
			return fmt.Errorf("unmarshal YAML %s failed: %w", opt.DataId, err)
		}
	}

	return nil
}

// LoadService loads the service-owned configuration followed by the shared
// infrastructure configuration. Loading infra last makes it the authoritative
// source during a gradual migration from duplicated service fields.
func LoadService(dataID string, serviceTarget any, infraTarget ...any) error {
	if len(infraTarget) > 1 {
		return fmt.Errorf("service %s accepts at most one infra target", dataID)
	}

	options := []LoadOption{Service(dataID, serviceTarget)}
	if len(infraTarget) == 1 {
		options = append(options, Infra(infraTarget[0]))
	}

	if err := Load(options...); err != nil {
		return err
	}

	if applier, ok := serviceTarget.(interface{ ApplyInfra() }); ok {
		applier.ApplyInfra()
	}

	return nil
}

func MustLoadService(dataID string, serviceTarget any, infraTarget ...any) {
	if err := LoadService(dataID, serviceTarget, infraTarget...); err != nil {
		panic(err)
	}
}

func newNacosClient() (config_client.IConfigClient, error) {
	serverConfigs := []constant.ServerConfig{
		{
			IpAddr: os.Getenv("NACOS_ADDR"), //
			Port:   8848,
		},
	}

	clientConfig := constant.ClientConfig{
		NamespaceId: os.Getenv("NACOS_NAMESPACE"), //muxi_fresh
		Username:    os.Getenv("NACOS_USERNAME"),
		Password:    os.Getenv("NACOS_PASSWORD"),
		TimeoutMs:   5000,
		LogLevel:    "warn",
	}

	return clients.NewConfigClient(
		vo.NacosClientParam{
			ServerConfigs: serverConfigs,
			ClientConfig:  &clientConfig,
		},
	)
}

func valueOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return fallback
}

func MustLoad(options ...LoadOption) {
	if err := Load(options...); err != nil {
		panic(err)
	}
}
