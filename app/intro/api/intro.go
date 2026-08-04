package main

import (
	"flag"
	"fmt"

	"MuXiFresh-Be-2.0/app/intro/api/internal/config"
	"MuXiFresh-Be-2.0/app/intro/api/internal/handler"
	"MuXiFresh-Be-2.0/app/intro/api/internal/svc"
	"MuXiFresh-Be-2.0/common/nacos"

	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "etc/intro-api.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	nacos.MustLoadService("intro-api", &c, &c.Infra)

	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()

	ctx := svc.NewServiceContext(c)
	handler.RegisterHandlers(server, ctx)

	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
