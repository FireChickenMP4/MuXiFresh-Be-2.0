package nacos

import (
	"encoding/json"
	"fmt"
	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/nacos-group/nacos-sdk-go/vo"
	"os"
)

type LoadOption struct {
	Group  string
	DataId string
	Target any
}

func Load(options ...LoadOption) error {
	client, err := newNacosClient()
	if err != nil {
		return err
	}

	for _, opt := range options {
		content, err := client.GetConfig(vo.ConfigParam{
			DataId: opt.DataId,
			Group:  opt.Group,
		})
		if err != nil {
			return fmt.Errorf("load %s failed: %w", opt.DataId, err)
		}

		if err := json.Unmarshal([]byte(content), opt.Target); err != nil {
			return fmt.Errorf("unmarshal %s failed: %w", opt.DataId, err)
		}
	}

	return nil
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
func MustLoad(options ...LoadOption) {
	if err := Load(options...); err != nil {
		panic(err)
	}
}
